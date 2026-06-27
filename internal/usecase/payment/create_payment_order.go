package payment

import (
	"context"
	"fmt"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
	domainprovider "github.com/mamahoos/airbar-finance/internal/domain/provider"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/zibal"
)

// ZibalGateway abstracts Zibal request/verify for use cases.
type ZibalGateway interface {
	Request(ctx context.Context, input zibal.RequestInput) (zibal.RequestResult, error)
	Verify(ctx context.Context, trackID string) (zibal.VerifyResult, error)
}

// CreatePaymentOrderInput is the application input for UC-09.
type CreatePaymentOrderInput struct {
	ShipmentID  string
	PayerUserID string
	Amount      int64
	SuccessURL  string
	FailureURL  string
	Description string
	AgreedPrice int64
}

// CreatePaymentOrder creates a shipment payment order and Zibal session.
type CreatePaymentOrder struct {
	orders      domainpayment.Repository
	escrowRepo  domainescrow.Repository
	events      domainprovider.Repository
	zibal       ZibalGateway
	callbackURL string
}

// NewCreatePaymentOrder creates the CreatePaymentOrder use case.
func NewCreatePaymentOrder(
	orders domainpayment.Repository,
	escrowRepo domainescrow.Repository,
	events domainprovider.Repository,
	zibalClient ZibalGateway,
	publicBaseURL string,
) *CreatePaymentOrder {
	return &CreatePaymentOrder{
		orders:      orders,
		escrowRepo:  escrowRepo,
		events:      events,
		zibal:       zibalClient,
		callbackURL: CallbackURL(publicBaseURL),
	}
}

// Execute validates escrow preconditions, requests Zibal, and persists the order.
func (uc *CreatePaymentOrder) Execute(ctx context.Context, input CreatePaymentOrderInput) (*domainpayment.Order, error) {
	if input.ShipmentID == "" || input.PayerUserID == "" || input.Amount <= 0 {
		return nil, domainpayment.ErrInvalidInput
	}
	if input.SuccessURL == "" || input.FailureURL == "" {
		return nil, domainpayment.ErrInvalidInput
	}
	if input.AgreedPrice > 0 && input.AgreedPrice != input.Amount {
		return nil, domainpayment.ErrAmountMismatch
	}

	escrow, err := uc.escrowRepo.GetByShipmentID(ctx, input.ShipmentID)
	if err != nil {
		return nil, err
	}
	if !escrow.Status.CanFund() {
		return nil, domainpayment.ErrEscrowNotReady
	}
	if escrow.PayerUserID != input.PayerUserID {
		return nil, domainescrow.ErrPayerMismatch
	}
	if escrow.Amount != input.Amount {
		return nil, domainpayment.ErrAmountMismatch
	}

	order := &domainpayment.Order{
		ShipmentID:  input.ShipmentID,
		PayerUserID: input.PayerUserID,
		Purpose:     domainpayment.PurposeShipment,
		Amount:      input.Amount,
		Status:      domainpayment.StatusPending,
		SuccessURL:  input.SuccessURL,
		FailureURL:  input.FailureURL,
		Description: input.Description,
		AgreedPrice: input.AgreedPrice,
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
