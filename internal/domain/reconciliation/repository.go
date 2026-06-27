package reconciliation

import "context"

// Repository persists reconciliation runs.
type Repository interface {
	Create(ctx context.Context, run *Run) error
	GetByID(ctx context.Context, id string) (*Run, error)
	List(ctx context.Context) ([]Run, error)
}
