package credit

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

// EnsureForLines lazily creates promo credit accounts for user credit lines in a journal.
func EnsureForLines(ctx context.Context, ensurer *EnsureCreditAccount, lines []domainledger.EntryLine) error {
	if ensurer == nil {
		return nil
	}

	seen := make(map[string]struct{})
	for _, line := range lines {
		userID, ok := domainledger.ParseUserIDFromPromoCreditAccount(line.AccountCode)
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
