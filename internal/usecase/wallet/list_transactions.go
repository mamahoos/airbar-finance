package wallet

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainwallet "github.com/mamahoos/airbar-finance/internal/domain/wallet"
)

// LedgerHistoryReader lists account entries from the ledger SSOT.
type LedgerHistoryReader interface {
	ListByAccount(ctx context.Context, accountCode domainledger.AccountCode) ([]domainledger.AccountEntry, error)
}

// ListWalletTransactions returns wallet movements from ledger journals (UC-15).
type ListWalletTransactions struct {
	ledger LedgerHistoryReader
}

// NewListWalletTransactions creates the ListWalletTransactions use case.
func NewListWalletTransactions(ledger LedgerHistoryReader) *ListWalletTransactions {
	return &ListWalletTransactions{ledger: ledger}
}

// Execute lists transactions for a user wallet account.
func (uc *ListWalletTransactions) Execute(ctx context.Context, userID, currency string) ([]domainwallet.Transaction, error) {
	if userID == "" {
		return nil, domainwallet.ErrInvalidInput
	}

	if _, err := NormalizeCurrency(currency); err != nil {
		return nil, err
	}

	entries, err := uc.ledger.ListByAccount(ctx, domainledger.UserWalletAccount(userID))
	if err != nil {
		return nil, err
	}

	items := make([]domainwallet.Transaction, len(entries))
	for i, entry := range entries {
		items[i] = domainwallet.Transaction{
			JournalID:   entry.JournalID,
			RefType:     string(entry.RefType),
			RefID:       entry.RefID,
			Description: entry.Description,
			Debit:       entry.Debit,
			Credit:      entry.Credit,
			CreatedAt:   entry.CreatedAt,
		}
	}
	return items, nil
}
