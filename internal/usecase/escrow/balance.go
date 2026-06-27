package escrow

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

// LedgerBalanceReader reads ledger sums for escrow balance derivation.
type LedgerBalanceReader interface {
	SumByAccount(ctx context.Context, accountCode domainledger.AccountCode) (debit int64, credit int64, err error)
}

// EscrowBalance returns the current escrow liability balance from ledger SSOT.
func EscrowBalance(ctx context.Context, ledger LedgerBalanceReader, shipmentID string) (int64, error) {
	debit, credit, err := ledger.SumByAccount(ctx, domainledger.ShipmentEscrowAccount(shipmentID))
	if err != nil {
		return 0, err
	}
	return credit - debit, nil
}
