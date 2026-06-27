package wallet

import (
	"context"
	"testing"
	"time"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

func TestGetWalletReturnsBalanceAndAccountCode(t *testing.T) {
	uc := NewGetWallet(NewGetBalance(mockBalanceReader{debit: 1000, credit: 6000}))

	wallet, err := uc.Execute(context.Background(), "user-1", "")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if wallet.Balance != 5000 {
		t.Fatalf("balance = %d, want 5000", wallet.Balance)
	}
	if wallet.Currency != "IRT" {
		t.Fatalf("currency = %q, want IRT", wallet.Currency)
	}
	if wallet.AccountCode != "USER:user-1:IRT:WALLET_LIABILITY" {
		t.Fatalf("account_code = %q", wallet.AccountCode)
	}
}

func TestGetWalletRejectsUnsupportedCurrency(t *testing.T) {
	uc := NewGetWallet(NewGetBalance(mockBalanceReader{}))

	_, err := uc.Execute(context.Background(), "user-1", "USD")
	if err == nil {
		t.Fatal("expected unsupported currency error")
	}
}

type mockHistoryReader struct {
	entries []domainledger.AccountEntry
}

func (m mockHistoryReader) ListByAccount(_ context.Context, _ domainledger.AccountCode) ([]domainledger.AccountEntry, error) {
	return m.entries, nil
}

func TestListWalletTransactionsMapsLedgerEntries(t *testing.T) {
	createdAt := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	uc := NewListWalletTransactions(mockHistoryReader{
		entries: []domainledger.AccountEntry{
			{
				JournalID:   "journal-1",
				RefType:     domainledger.RefTypeWalletTopup,
				RefID:       "user-1:topup",
				Description: "topup",
				Debit:       0,
				Credit:      10000,
				CreatedAt:   createdAt,
			},
		},
	})

	items, err := uc.Execute(context.Background(), "user-1", "IRT")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].Credit != 10000 {
		t.Fatalf("credit = %d, want 10000", items[0].Credit)
	}
}
