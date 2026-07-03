package credit

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

// BalanceReader reads ledger sums for promo credit balance derivation.
type BalanceReader interface {
	SumByAccount(ctx context.Context, accountCode domainledger.AccountCode) (debit int64, credit int64, err error)
}

// GetBalance returns promo credit balance from ledger SSOT: credit - debit on PROMO_CREDIT_LIABILITY.
type GetBalance struct {
	ledger BalanceReader
}

// NewGetBalance creates the GetBalance use case.
func NewGetBalance(ledger BalanceReader) *GetBalance {
	return &GetBalance{ledger: ledger}
}

// Execute returns non-withdrawable promo credit balance in rials.
func (uc *GetBalance) Execute(ctx context.Context, userID string) (int64, error) {
	debit, credit, err := uc.ledger.SumByAccount(ctx, domainledger.UserPromoCreditAccount(userID))
	if err != nil {
		return 0, err
	}
	return credit - debit, nil
}
