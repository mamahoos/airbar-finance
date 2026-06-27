package idempotency

import "context"

// Repository persists idempotency records.
type Repository interface {
	TryBeginProcessing(ctx context.Context, record *Record) (acquired bool, existing *Record, err error)
	GetByKey(ctx context.Context, key string) (*Record, error)
	Complete(ctx context.Context, key string, snapshot map[string]any) error
	DeleteProcessing(ctx context.Context, key string) error
}
