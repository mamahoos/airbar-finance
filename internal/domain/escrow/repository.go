package escrow

import "context"

// Repository persists escrow aggregates.
type Repository interface {
	Create(ctx context.Context, escrow *Escrow) error
	GetByShipmentID(ctx context.Context, shipmentID string) (*Escrow, error)
	Update(ctx context.Context, escrow *Escrow) error
}
