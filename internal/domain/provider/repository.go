package provider

import "context"

// Repository persists provider audit events.
type Repository interface {
	Create(ctx context.Context, event *Event) error
	CountByPaymentOrderID(ctx context.Context, paymentOrderID string) (int64, error)
	List(ctx context.Context, filter ListFilter) ([]Event, int64, error)
}

// ListFilter scopes provider event history for operational dashboards.
type ListFilter struct {
	Provider       string
	EventType      EventType
	PaymentOrderID string
	Page           int
	Limit          int
}
