package ledger

import "context"

// Repository persists journals and supports balance queries.
type Repository interface {
	CreateJournal(ctx context.Context, journal *Journal) error
	SumByAccount(ctx context.Context, accountCode AccountCode) (debit int64, credit int64, err error)
}
