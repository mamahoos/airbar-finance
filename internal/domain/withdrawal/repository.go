package withdrawal

import "context"

// Repository persists withdrawal aggregates.
type Repository interface {
	Create(ctx context.Context, withdrawal *Withdrawal) error
	GetByID(ctx context.Context, id string) (*Withdrawal, error)
	List(ctx context.Context, userID string, status Status) ([]Withdrawal, error)
	Update(ctx context.Context, withdrawal *Withdrawal) error
}
