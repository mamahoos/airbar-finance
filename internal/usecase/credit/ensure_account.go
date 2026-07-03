package credit

import (
	"context"

	domaincredit "github.com/mamahoos/airbar-finance/internal/domain/credit"
)

// EnsureCreditAccount lazily registers a promo credit account for the user.
type EnsureCreditAccount struct {
	repo domaincredit.Repository
}

// NewEnsureCreditAccount creates the use case.
func NewEnsureCreditAccount(repo domaincredit.Repository) *EnsureCreditAccount {
	return &EnsureCreditAccount{repo: repo}
}

// Execute registers the promo credit account if missing.
func (uc *EnsureCreditAccount) Execute(ctx context.Context, userID string) (*domaincredit.Account, error) {
	return uc.repo.EnsureAccount(ctx, userID)
}
