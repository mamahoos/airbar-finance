package provider

import "context"

// Repository persists provider audit events.
type Repository interface {
	Create(ctx context.Context, event *Event) error
}
