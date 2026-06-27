package wallet

import (
	"context"
	"testing"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

type mockBalanceReader struct {
	debit  int64
	credit int64
}

func (m mockBalanceReader) SumByAccount(_ context.Context, _ domainledger.AccountCode) (int64, int64, error) {
	return m.debit, m.credit, nil
}

func TestGetBalanceFromLedgerSums(t *testing.T) {
	uc := NewGetBalance(mockBalanceReader{debit: 2000, credit: 7000})

	balance, err := uc.Execute(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if balance != 5000 {
		t.Fatalf("balance = %d, want 5000", balance)
	}
}

func TestGetBalanceEmptyWallet(t *testing.T) {
	uc := NewGetBalance(mockBalanceReader{})

	balance, err := uc.Execute(context.Background(), "user-2")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if balance != 0 {
		t.Fatalf("balance = %d, want 0", balance)
	}
}
