package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	paymentuc "github.com/mamahoos/airbar-finance/internal/usecase/payment"
)

type callbackStub struct {
	result paymentuc.HandleCallbackResult
	err    error
	last   paymentuc.HandleCallbackInput
}

func (s *callbackStub) Execute(_ context.Context, input paymentuc.HandleCallbackInput) (paymentuc.HandleCallbackResult, error) {
	s.last = input
	return s.result, s.err
}

func TestZibalCallbackHandlerRedirectsSuccess(t *testing.T) {
	stub := &callbackStub{
		result: paymentuc.HandleCallbackResult{RedirectURL: "https://app/success"},
	}
	handler := &ZibalCallbackHandler{handleCallback: stub}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/zibal/callback?trackId=123&success=1", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "https://app/success" {
		t.Fatalf("location = %q", got)
	}
	if stub.last.TrackID != "123" || stub.last.Success != "1" {
		t.Fatalf("unexpected input: %+v", stub.last)
	}
}
