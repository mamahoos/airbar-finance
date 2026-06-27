package ledger

import (
	"context"
	"time"
)

// Repository persists journals and supports balance queries.
type Repository interface {
	CreateJournal(ctx context.Context, journal *Journal) error
	SumByAccount(ctx context.Context, accountCode AccountCode) (debit int64, credit int64, err error)
	SumGlobal(ctx context.Context) (debit int64, credit int64, err error)
	SumByAccountLike(ctx context.Context, pattern string) (debit int64, credit int64, err error)
	ListByAccount(ctx context.Context, accountCode AccountCode) ([]AccountEntry, error)
}

// AccountEntry is a ledger line with journal metadata for wallet history.
type AccountEntry struct {
	JournalID   string
	RefType     RefType
	RefID       string
	Description string
	Debit       int64
	Credit      int64
	CreatedAt   time.Time
}
