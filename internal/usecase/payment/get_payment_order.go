package payment

import (
	"context"

	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
)

// GetPaymentOrder loads a payment order by id (UC-10).
type GetPaymentOrder struct {
	orders domainpayment.Repository
}

// NewGetPaymentOrder creates the GetPaymentOrder use case.
func NewGetPaymentOrder(orders domainpayment.Repository) *GetPaymentOrder {
	return &GetPaymentOrder{orders: orders}
}

// Execute returns the payment order.
func (uc *GetPaymentOrder) Execute(ctx context.Context, orderID string) (*domainpayment.Order, error) {
	if orderID == "" {
		return nil, domainpayment.ErrInvalidInput
	}
	return uc.orders.GetByID(ctx, orderID)
}
