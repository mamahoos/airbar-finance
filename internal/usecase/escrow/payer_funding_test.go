package escrow

import (
	"errors"
	"testing"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

func TestResolvePayerFundingSplitPromoOnly(t *testing.T) {
	split, err := ResolvePayerFundingSplit(10000, 15000, 0)
	if err != nil {
		t.Fatalf("ResolvePayerFundingSplit() error = %v", err)
	}
	if split.PromoCredit != 10000 || split.Wallet != 0 {
		t.Fatalf("split = %+v, want promo=10000 wallet=0", split)
	}
	if FundingSourceForSplit(split) != domainescrow.FundingSourcePromoCredit {
		t.Fatalf("funding source = %q", FundingSourceForSplit(split))
	}
}

func TestResolvePayerFundingSplitMixed(t *testing.T) {
	split, err := ResolvePayerFundingSplit(10000, 3000, 9000)
	if err != nil {
		t.Fatalf("ResolvePayerFundingSplit() error = %v", err)
	}
	if split.PromoCredit != 3000 || split.Wallet != 7000 {
		t.Fatalf("split = %+v, want promo=3000 wallet=7000", split)
	}
	if FundingSourceForSplit(split) != domainescrow.FundingSourceMixed {
		t.Fatalf("funding source = %q", FundingSourceForSplit(split))
	}
}

func TestResolvePayerFundingSplitInsufficient(t *testing.T) {
	_, err := ResolvePayerFundingSplit(10000, 2000, 3000)
	if !errors.Is(err, domainescrow.ErrInsufficientPayerFunds) {
		t.Fatalf("error = %v, want ErrInsufficientPayerFunds", err)
	}
}

func TestRefundAllocationPromoFirst(t *testing.T) {
	promo, wallet := RefundAllocation(8000, 5000)
	if promo != 5000 || wallet != 3000 {
		t.Fatalf("promo=%d wallet=%d, want 5000/3000", promo, wallet)
	}
}

func TestBuildEscrowPayJournalLinesBalanced(t *testing.T) {
	lines := BuildEscrowPayJournalLines("user-1", "sh-1", PayerFundingSplit{PromoCredit: 3000, Wallet: 7000})
	var debit int64
	var credit int64
	for _, line := range lines {
		debit += line.Debit
		credit += line.Credit
	}
	if debit != credit || debit != 10000 {
		t.Fatalf("journal not balanced: debit=%d credit=%d", debit, credit)
	}
	if err := domainledger.ValidateLines(lines); err != nil {
		t.Fatalf("ValidateLines() error = %v", err)
	}
}
