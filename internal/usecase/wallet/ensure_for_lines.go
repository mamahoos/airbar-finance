package wallet

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

// EnsureForLines lazily creates wallet accounts for user wallet lines in a journal.
func EnsureForLines(ctx context.Context, ensurer *EnsureWalletAccount, lines []domainledger.EntryLine) error {
	if ensurer == nil {
		return nil
	}

	seen := make(map[string]struct{})
	for _, line := range lines {
		userID, ok := domainledger.ParseUserIDFromWalletAccount(line.AccountCode)
		if !ok {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		if _, err := ensurer.Execute(ctx, userID); err != nil {
			return err
		}
	}
	return nil
}
