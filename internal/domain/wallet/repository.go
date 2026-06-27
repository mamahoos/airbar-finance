package wallet

import "context"

// Repository manages wallet account registration (lazy create).
type Repository interface {
	EnsureAccount(ctx context.Context, userID string) (*Account, error)
}
