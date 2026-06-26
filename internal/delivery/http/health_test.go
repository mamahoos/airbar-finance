package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubChecker struct {
	ready bool
}

func (s stubChecker) Ready(_ context.Context) bool {
	return s.ready
}

func TestHealthHandlerReady(t *testing.T) {
	handler := NewHealthHandler(stubChecker{ready: true})
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q, want ok", rec.Body.String())
	}
}

func TestHealthHandlerNotReady(t *testing.T) {
	handler := NewHealthHandler(stubChecker{ready: false})
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if rec.Body.String() != "not ready" {
		t.Fatalf("body = %q, want not ready", rec.Body.String())
	}
}
