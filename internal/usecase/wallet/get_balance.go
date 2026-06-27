package wallet

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

// BalanceReader reads ledger sums for wallet balance derivation.
type BalanceReader interface {
	SumByAccount(ctx context.Context, accountCode domainledger.AccountCode) (debit int64, credit int64, err error)
}

// GetBalance returns wallet balance from ledger SSOT: credit - debit on WALLET_LIABILITY.
type GetBalance struct {
	ledger BalanceReader
}

// NewGetBalance creates the GetBalance use case.
func NewGetBalance(ledger BalanceReader) *GetBalance {
	return &GetBalance{ledger: ledger}
}

// Execute returns spendable wallet balance in rials (integer).
func (uc *GetBalance) Execute(ctx context.Context, userID string) (int64, error) {
	debit, credit, err := uc.ledger.SumByAccount(ctx, domainledger.UserWalletAccount(userID))
	if err != nil {
		return 0, err
	}
	return credit - debit, nil
}
