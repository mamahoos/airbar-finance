package treasury

import (
	"context"
	"testing"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

type stubLedger struct {
	byAccount map[string]struct{ debit, credit int64 }
	byPattern map[string]struct{ debit, credit int64 }
}

func (s *stubLedger) SumByAccount(_ context.Context, code domainledger.AccountCode) (int64, int64, error) {
	v, ok := s.byAccount[code.String()]
	if !ok {
		return 0, 0, nil
	}
	return v.debit, v.credit, nil
}

func (s *stubLedger) SumByAccountLike(_ context.Context, pattern string) (int64, int64, error) {
	v, ok := s.byPattern[pattern]
	if !ok {
		return 0, 0, nil
	}
	return v.debit, v.credit, nil
}

func TestGetTreasurySummary(t *testing.T) {
	ledger := &stubLedger{
		byAccount: map[string]struct{ debit, credit int64 }{
			domainledger.AccountIRPSPClearing.String():    {debit: 5000, credit: 1000},
			domainledger.AccountIRBankMain.String():       {debit: 0, credit: 0},
			domainledger.AccountIRPayoutClearing.String(): {debit: 0, credit: 2000},
			domainledger.AccountAirbarFeeRevenue.String():  {debit: 0, credit: 300},
		},
		byPattern: map[string]struct{ debit, credit int64 }{
			domainledger.WalletAccountLikePattern(): {debit: 500, credit: 2500},
			domainledger.EscrowAccountLikePattern(): {debit: 100, credit: 4000},
		},
	}

	uc := NewGetTreasurySummary(ledger)
	summary, err := uc.Execute(context.Background(), "IRT")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if summary.Accounts[domainledger.AccountIRPSPClearing.String()] != 4000 {
		t.Fatalf("PSP clearing = %d, want 4000", summary.Accounts[domainledger.AccountIRPSPClearing.String()])
	}
	if summary.Accounts["AGGREGATE_WALLET_LIABILITY"] != 2000 {
		t.Fatalf("wallet aggregate = %d, want 2000", summary.Accounts["AGGREGATE_WALLET_LIABILITY"])
	}
	if summary.Accounts["AGGREGATE_ESCROW_LIABILITY"] != 3900 {
		t.Fatalf("escrow aggregate = %d, want 3900", summary.Accounts["AGGREGATE_ESCROW_LIABILITY"])
	}
}

func TestGetTreasurySummaryUnsupportedCurrency(t *testing.T) {
	uc := NewGetTreasurySummary(&stubLedger{})
	_, err := uc.Execute(context.Background(), "USD")
	if err != ErrUnsupportedCurrency {
		t.Fatalf("Execute() error = %v, want ErrUnsupportedCurrency", err)
	}
}
