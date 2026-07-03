package credit

import (
	"context"
	"time"
)

// Repository manages promo credit accounts and grant records.
type Repository interface {
	EnsureAccount(ctx context.Context, userID string) (*Account, error)
	CreateGrant(ctx context.Context, grant *Grant) error
	GetGrantByID(ctx context.Context, id string) (*Grant, error)
	MarkReversed(ctx context.Context, id string, reversedAt time.Time, reverseReason, reversedBy string) error
	ListGrantsByUserID(ctx context.Context, userID string, limit, offset int) ([]Grant, error)
}
