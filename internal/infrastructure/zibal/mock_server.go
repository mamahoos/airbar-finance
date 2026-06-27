package zibal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
)

// MockServer hosts fake Zibal request/verify endpoints for tests.
type MockServer struct {
	Server *httptest.Server
	mu     sync.Mutex
	orders map[string]mockOrder
}

type mockOrder struct {
	amount int64
	paid   bool
}

// NewMockServer creates an httptest Zibal gateway.
func NewMockServer() *MockServer {
	mock := &MockServer{orders: make(map[string]mockOrder)}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/request", mock.handleRequest)
	mux.HandleFunc("/v1/verify", mock.handleVerify)
	mock.Server = httptest.NewServer(mux)
	return mock
}

func (m *MockServer) Close() {
	m.Server.Close()
}

func (m *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Amount  int64  `json:"amount"`
		OrderID string `json:"orderId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trackID := strconv.FormatInt(int64(len(m.orders)+1), 10)
	m.mu.Lock()
	m.orders[trackID] = mockOrder{amount: body.Amount}
	m.mu.Unlock()

	trackNum, _ := strconv.ParseInt(trackID, 10, 64)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"result":  100,
		"trackId": trackNum,
		"message": "success",
	})
}

func (m *MockServer) handleVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TrackID string `json:"trackId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	order, ok := m.orders[body.TrackID]
	if ok {
		order.paid = true
		m.orders[body.TrackID] = order
	}
	m.mu.Unlock()

	if !ok {
		_ = json.NewEncoder(w).Encode(map[string]any{"result": 201, "message": "not found"})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"result":  100,
		"amount":  order.amount,
		"message": "success",
	})
}

// MarkPaid marks a trackId as paid in the mock (simulates user completing checkout).
func (m *MockServer) MarkPaid(trackID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if order, ok := m.orders[trackID]; ok {
		order.paid = true
		m.orders[trackID] = order
	}
}

// Client returns a Zibal client wired to this mock server.
func (m *MockServer) Client() *Client {
	return (&Client{
		httpClient: m.Server.Client(),
		merchant:   "zibal",
		baseURL:    m.Server.URL,
	})
}

// Request is a convenience wrapper for tests.
func (m *MockServer) Request(ctx context.Context, input RequestInput) (RequestResult, error) {
	return m.Client().Request(ctx, input)
}

// Verify is a convenience wrapper for tests.
func (m *MockServer) Verify(ctx context.Context, trackID string) (VerifyResult, error) {
	return m.Client().Verify(ctx, trackID)
}
