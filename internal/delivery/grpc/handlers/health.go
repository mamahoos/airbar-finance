package handlers

import (
	"context"

	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/health"
)

// HealthHandler implements FinanceHealthService.CheckReady (UC-24).
type HealthHandler struct {
	financev1.UnimplementedFinanceHealthServiceServer
	checker *health.Checker
}

// NewHealthHandler creates a gRPC health handler.
func NewHealthHandler(checker *health.Checker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// CheckReady reports readiness based on Postgres and Redis connectivity.
func (h *HealthHandler) CheckReady(ctx context.Context, _ *financev1.HealthCheckRequest) (*financev1.HealthCheckResponse, error) {
	return &financev1.HealthCheckResponse{Ready: h.checker.Ready(ctx)}, nil
}
