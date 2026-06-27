package payment

import (
	"context"
	"fmt"

	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
	domainprovider "github.com/mamahoos/airbar-finance/internal/domain/provider"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/zibal"
)

// CreateWalletTopupOrderInput is the application input for UC-12.
type CreateWalletTopupOrderInput struct {
	UserID      string
	Amount      int64
	SuccessURL  string
	FailureURL  string
	Description string
}

// CreateWalletTopupOrder creates a wallet topup payment order and Zibal session.
type CreateWalletTopupOrder struct {
	orders      domainpayment.Repository
	events      domainprovider.Repository
	zibal       ZibalGateway
	callbackURL string
}

// NewCreateWalletTopupOrder creates the CreateWalletTopupOrder use case.
func NewCreateWalletTopupOrder(
	orders domainpayment.Repository,
	events domainprovider.Repository,
	zibalClient ZibalGateway,
	publicBaseURL string,
) *CreateWalletTopupOrder {
	return &CreateWalletTopupOrder{
		orders:      orders,
		events:      events,
		zibal:       zibalClient,
		callbackURL: CallbackURL(publicBaseURL),
	}
}

// Execute requests Zibal and persists a WALLET_TOPUP order.
func (uc *CreateWalletTopupOrder) Execute(ctx context.Context, input CreateWalletTopupOrderInput) (*domainpayment.Order, error) {
	if input.UserID == "" || input.Amount <= 0 {
		return nil, domainpayment.ErrInvalidInput
	}
	if input.SuccessURL == "" || input.FailureURL == "" {
		return nil, domainpayment.ErrInvalidInput
	}

	order := &domainpayment.Order{
		PayerUserID: input.UserID,
		Purpose:     domainpayment.PurposeWalletTopup,
		Amount:      input.Amount,
		Status:      domainpayment.StatusPending,
		SuccessURL:  input.SuccessURL,
		FailureURL:  input.FailureURL,
		Description: input.Description,
	}
	if err := uc.orders.Create(ctx, order); err != nil {
		return nil, err
	}

	result, err := uc.zibal.Request(ctx, zibal.RequestInput{
		Amount:      order.Amount,
		CallbackURL: uc.callbackURL,
		Description: order.Description,
		OrderID:     order.ID,
	})
	if err != nil {
		order.Status = domainpayment.StatusFailed
		_ = uc.orders.Update(ctx, order)
		return nil, err
	}

	order.Authority = result.TrackID
	order.RedirectURL = result.RedirectURL
	if err := uc.orders.Update(ctx, order); err != nil {
		return nil, err
	}

	payload, hash, err := HashPayload(map[string]any{
		"trackId": result.TrackID,
		"orderId": order.ID,
	})
	if err != nil {
		return nil, err
	}
	_ = uc.events.Create(ctx, &domainprovider.Event{
		Provider:       domainprovider.ProviderZibal,
		EventType:      domainprovider.EventTypeRequest,
		PaymentOrderID: order.ID,
		Payload:        payload,
		PayloadHash:    hash,
		IdempotencyKey: fmt.Sprintf("zibal:request:%s", order.ID),
		Processed:      true,
	})

	return order, nil
}
