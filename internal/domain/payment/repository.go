package payment

import "context"

// Repository persists payment orders.
type Repository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	GetByAuthority(ctx context.Context, authority string) (*Order, error)
	Update(ctx context.Context, order *Order) error
}
