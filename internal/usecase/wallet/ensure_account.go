package wallet

import (
	"context"

	domainwallet "github.com/mamahoos/airbar-finance/internal/domain/wallet"
)

// EnsureWalletAccount lazily registers a wallet account for the user.
type EnsureWalletAccount struct {
	repo domainwallet.Repository
}

// NewEnsureWalletAccount creates the use case.
func NewEnsureWalletAccount(repo domainwallet.Repository) *EnsureWalletAccount {
	return &EnsureWalletAccount{repo: repo}
}

// Execute registers the wallet if missing and returns the account metadata.
func (uc *EnsureWalletAccount) Execute(ctx context.Context, userID string) (*domainwallet.Account, error) {
	return uc.repo.EnsureAccount(ctx, userID)
}
