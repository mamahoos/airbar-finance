package payment

import (
	"context"
	"fmt"

	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
	domainprovider "github.com/mamahoos/airbar-finance/internal/domain/provider"
)

// HandleCallbackInput is input for the Zibal HTTP callback.
type HandleCallbackInput struct {
	TrackID string
	Success string
}

// HandleCallbackResult is the redirect target after callback processing.
type HandleCallbackResult struct {
	RedirectURL string
	Order       *domainpayment.Order
}

// HandleCallback processes Zibal GET callback and returns redirect URL.
type HandleCallback struct {
	verify    *VerifyOrder
	failOrder *FailPaymentOrder
	events    domainprovider.Repository
}

// NewHandleCallback creates the callback handler use case.
func NewHandleCallback(
	verify *VerifyOrder,
	failOrder *FailPaymentOrder,
	events domainprovider.Repository,
) *HandleCallback {
	return &HandleCallback{
		verify:    verify,
		failOrder: failOrder,
		events:    events,
	}
}

// Execute verifies or fails the order and picks success/failure redirect URL.
func (uc *HandleCallback) Execute(ctx context.Context, input HandleCallbackInput) (HandleCallbackResult, error) {
	if input.TrackID == "" {
		return HandleCallbackResult{}, domainpayment.ErrInvalidInput
	}

	payload, hash, err := HashPayload(map[string]any{
		"trackId": input.TrackID,
		"success": input.Success,
	})
	if err != nil {
		return HandleCallbackResult{}, err
	}
	_ = uc.events.Create(ctx, &domainprovider.Event{
		Provider:       domainprovider.ProviderZibal,
		EventType:      domainprovider.EventTypeCallback,
		Payload:        payload,
		PayloadHash:    hash,
		IdempotencyKey: fmt.Sprintf("zibal:callback:%s:%s:received", input.TrackID, input.Success),
		Processed:      false,
	})

	if input.Success != "1" {
		order, err := uc.failOrder.Execute(ctx, input.TrackID, input.Success)
		if err != nil {
			return HandleCallbackResult{}, err
		}
		return HandleCallbackResult{RedirectURL: order.FailureURL, Order: order}, nil
	}

	order, err := uc.verify.ExecuteByAuthority(ctx, input.TrackID)
	if err != nil {
		if failed, failErr := uc.failOrder.Execute(ctx, input.TrackID, input.Success); failErr == nil {
			return HandleCallbackResult{RedirectURL: failed.FailureURL, Order: failed}, nil
		}
		return HandleCallbackResult{}, err
	}

	return HandleCallbackResult{RedirectURL: order.SuccessURL, Order: order}, nil
}
