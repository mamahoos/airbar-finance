package http

import (
	"context"
	"net/http"
)

// ReadinessChecker reports whether the service is ready to accept traffic.
type ReadinessChecker interface {
	Ready(ctx context.Context) bool
}

// HealthHandler serves GET /health/ready.
type HealthHandler struct {
	checker ReadinessChecker
}

// NewHealthHandler creates an HTTP readiness handler.
func NewHealthHandler(checker ReadinessChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// ServeHTTP returns 200 when ready, 503 otherwise.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.checker.Ready(r.Context()) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}

	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("not ready"))
}
