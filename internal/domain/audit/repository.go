package audit

import "context"

// Repository persists finance audit events.
type Repository interface {
	Create(ctx context.Context, event *Event) error
	CountByAggregate(ctx context.Context, aggregateType, aggregateID string) (int64, error)
}
