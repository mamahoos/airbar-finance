//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	domaincredit "github.com/mamahoos/airbar-finance/internal/domain/credit"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	credituc "github.com/mamahoos/airbar-finance/internal/usecase/credit"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	treasuryuc "github.com/mamahoos/airbar-finance/internal/usecase/treasury"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

func TestCreditGrantAndReverseIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close()

	ledgerRepo := NewLedgerRepository(pool)
	creditRepo := NewCreditRepository(pool)
	ensureCredit := credituc.NewEnsureCreditAccount(creditRepo)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, nil)
	getBalance := credituc.NewGetBalance(ledgerRepo)
	getTreasury := treasuryuc.NewGetTreasurySummary(ledgerRepo)
	grantCredit := credituc.NewGrantCredit(pool, creditRepo, ensureCredit, postJournal, audituc.NewEmitter(NewFinanceEventRepository(pool)))
	reverseCredit := credituc.NewReverseCreditGrant(pool, creditRepo, postJournal, audituc.NewEmitter(NewFinanceEventRepository(pool)))

	userID := fmt.Sprintf("credit-user-%s", uuid.NewString()[:8])
	adminID := fmt.Sprintf("admin-%s", uuid.NewString()[:8])
	amount := int64(25000)

	beforeTreasury, err := getTreasury.Execute(ctx, "IRT")
	if err != nil {
		t.Fatalf("GetTreasurySummary() before error = %v", err)
	}

	grant, err := grantCredit.Execute(ctx, credituc.GrantCreditInput{
		UserID:         userID,
		Amount:         amount,
		Reason:         "welcome bonus",
		GrantedBy:      adminID,
		IdempotencyKey: uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("GrantCredit() error = %v", err)
	}
	if grant.Status != domaincredit.StatusActive {
		t.Fatalf("grant status = %q, want ACTIVE", grant.Status)
	}

	balance, err := getBalance.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if balance != amount {
		t.Fatalf("balance after grant = %d, want %d", balance, amount)
	}

	afterGrantTreasury, err := getTreasury.Execute(ctx, "IRT")
	if err != nil {
		t.Fatalf("GetTreasurySummary() after grant error = %v", err)
	}
	if delta := afterGrantTreasury.Accounts["AGGREGATE_PROMO_CREDIT_LIABILITY"] - beforeTreasury.Accounts["AGGREGATE_PROMO_CREDIT_LIABILITY"]; delta != amount {
		t.Fatalf("promo liability delta = %d, want %d", delta, amount)
	}
	if delta := afterGrantTreasury.Accounts[string(domainledger.AccountAirbarPromoExpense)] - beforeTreasury.Accounts[string(domainledger.AccountAirbarPromoExpense)]; delta != amount {
		t.Fatalf("promo expense delta = %d, want %d", delta, amount)
	}

	reversed, err := reverseCredit.Execute(ctx, credituc.ReverseCreditInput{
		GrantID:       grant.ID,
		ReverseReason: "manual correction",
		ReversedBy:    adminID,
	})
	if err != nil {
		t.Fatalf("ReverseCreditGrant() error = %v", err)
	}
	if reversed.Status != domaincredit.StatusReversed {
		t.Fatalf("reversed status = %q, want REVERSED", reversed.Status)
	}

	balanceAfterReverse, err := getBalance.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("GetBalance() after reverse error = %v", err)
	}
	if balanceAfterReverse != 0 {
		t.Fatalf("balance after reverse = %d, want 0", balanceAfterReverse)
	}
}
