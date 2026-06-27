package zibal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mamahoos/airbar-finance/internal/infrastructure/config"
)

const defaultBaseURL = "https://gateway.zibal.ir"

const resultSuccess = 100

// RequestInput is input for creating a Zibal payment session.
type RequestInput struct {
	Amount      int64
	CallbackURL string
	Description string
	OrderID     string
}

// RequestResult is the outcome of a Zibal request call.
type RequestResult struct {
	TrackID     string
	RedirectURL string
}

// VerifyResult is the outcome of a Zibal verify call.
type VerifyResult struct {
	Amount int64
}

// Client calls Zibal gateway APIs.
type Client struct {
	httpClient *http.Client
	merchant   string
	baseURL    string
}

// NewClient creates a Zibal HTTP client from config.
func NewClient(cfg config.Config) *Client {
	merchant := cfg.ZibalMerchant
	if cfg.ZibalSandbox {
		merchant = "zibal"
	}
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		merchant:   merchant,
		baseURL:    defaultBaseURL,
	}
}

// WithBaseURL returns a copy of the client pointed at a custom base URL (tests).
func (c *Client) WithBaseURL(baseURL string) *Client {
	clone := *c
	clone.baseURL = strings.TrimRight(baseURL, "/")
	return &clone
}

// Request creates a payment session and returns trackId + redirect URL.
func (c *Client) Request(ctx context.Context, input RequestInput) (RequestResult, error) {
	body := map[string]any{
		"merchant":    c.merchant,
		"amount":      input.Amount,
		"callbackUrl": input.CallbackURL,
		"description": input.Description,
		"orderId":     input.OrderID,
	}

	var resp apiResponse
	if err := c.post(ctx, "/v1/request", body, &resp); err != nil {
		return RequestResult{}, err
	}
	if resp.Result != resultSuccess {
		return RequestResult{}, fmt.Errorf("zibal request failed: result=%d message=%s", resp.Result, resp.Message)
	}

	trackID := fmt.Sprintf("%d", resp.TrackID)
	return RequestResult{
		TrackID:     trackID,
		RedirectURL: fmt.Sprintf("%s/start/%s", c.baseURL, trackID),
	}, nil
}

// Verify confirms a payment by trackId.
func (c *Client) Verify(ctx context.Context, trackID string) (VerifyResult, error) {
	body := map[string]any{
		"merchant": c.merchant,
		"trackId":  trackID,
	}

	var resp verifyResponse
	if err := c.post(ctx, "/v1/verify", body, &resp); err != nil {
		return VerifyResult{}, err
	}
	if resp.Result != resultSuccess {
		return VerifyResult{}, fmt.Errorf("zibal verify failed: result=%d message=%s", resp.Result, resp.Message)
	}

	return VerifyResult{Amount: resp.Amount}, nil
}

type apiResponse struct {
	Result  int    `json:"result"`
	TrackID int64  `json:"trackId"`
	Message string `json:"message"`
}

type verifyResponse struct {
	Result  int    `json:"result"`
	Amount  int64  `json:"amount"`
	Message string `json:"message"`
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("zibal http %d: %s", resp.StatusCode, string(raw))
	}
	return json.Unmarshal(raw, out)
}
