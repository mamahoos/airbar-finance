package http

import (
	"context"
	"net/http"

	paymentuc "github.com/mamahoos/airbar-finance/internal/usecase/payment"
)

type callbackProcessor interface {
	Execute(ctx context.Context, input paymentuc.HandleCallbackInput) (paymentuc.HandleCallbackResult, error)
}

// ZibalCallbackHandler serves GET /api/v1/zibal/callback.
type ZibalCallbackHandler struct {
	handleCallback callbackProcessor
}

// NewZibalCallbackHandler creates the Zibal callback HTTP handler.
func NewZibalCallbackHandler(handleCallback *paymentuc.HandleCallback) *ZibalCallbackHandler {
	return &ZibalCallbackHandler{handleCallback: handleCallback}
}

// ServeHTTP verifies payment and redirects the user browser.
func (h *ZibalCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	result, err := h.handleCallback.Execute(r.Context(), paymentuc.HandleCallbackInput{
		TrackID: r.URL.Query().Get("trackId"),
		Success: r.URL.Query().Get("success"),
	})
	if err != nil || result.RedirectURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("payment callback failed"))
		return
	}

	http.Redirect(w, r, result.RedirectURL, http.StatusFound)
}
