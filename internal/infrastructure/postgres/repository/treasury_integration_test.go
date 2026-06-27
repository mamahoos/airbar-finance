//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainrecon "github.com/mamahoos/airbar-finance/internal/domain/reconciliation"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	reconuc "github.com/mamahoos/airbar-finance/internal/usecase/reconciliation"
	treasuryuc "github.com/mamahoos/airbar-finance/internal/usecase/treasury"
)

func TestTreasuryAndReconciliationIntegration(t *testing.T) {
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
	reconRepo := NewReconciliationRepository(pool)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, nil)
	getTreasury := treasuryuc.NewGetTreasurySummary(ledgerRepo)

	before, err := getTreasury.Execute(ctx, "IRT")
	if err != nil {
		t.Fatalf("GetTreasurySummary() before error = %v", err)
	}

	userID := fmt.Sprintf("treasury-user-%s", uuid.NewString()[:8])
	shipmentID := fmt.Sprintf("treasury-sh-%s", uuid.NewString()[:8])
	amount := int64(1500)

	_, err = postJournal.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletTopup,
		RefID:   userID + ":topup:" + uuid.NewString()[:8],
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.AccountIRPSPClearing, Debit: amount, Credit: 0},
			{AccountCode: domainledger.UserWalletAccount(userID), Debit: 0, Credit: amount},
		},
	})
	if err != nil {
		t.Fatalf("PostJournal topup error = %v", err)
	}

	after, err := getTreasury.Execute(ctx, "IRT")
	if err != nil {
		t.Fatalf("GetTreasurySummary() after topup error = %v", err)
	}
	if delta := after.Accounts[string(domainledger.AccountIRPSPClearing)] - before.Accounts[string(domainledger.AccountIRPSPClearing)]; delta != amount {
		t.Fatalf("PSP clearing delta = %d, want %d", delta, amount)
	}
	if delta := after.Accounts["AGGREGATE_WALLET_LIABILITY"] - before.Accounts["AGGREGATE_WALLET_LIABILITY"]; delta != amount {
		t.Fatalf("wallet aggregate delta = %d, want %d", delta, amount)
	}

	runRecon := reconuc.NewRunReconciliation(ledgerRepo, reconRepo)
	run, err := runRecon.Execute(ctx)
	if err != nil {
		t.Fatalf("RunReconciliation() error = %v", err)
	}
	if run.Status != domainrecon.StatusPassed {
		t.Fatalf("reconciliation status = %q, want PASSED", run.Status)
	}

	listRuns := reconuc.NewListReconciliationRuns(reconRepo)
	runs, err := listRuns.Execute(ctx)
	if err != nil {
		t.Fatalf("ListReconciliationRuns() error = %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected reconciliation runs")
	}

	getRun := reconuc.NewGetReconciliationRun(reconRepo)
	loaded, err := getRun.Execute(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetReconciliationRun() error = %v", err)
	}
	if loaded.Status != domainrecon.StatusPassed {
		t.Fatalf("loaded status = %q, want PASSED", loaded.Status)
	}

	_, err = postJournal.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletToEscrow,
		RefID:   shipmentID + ":wallet-pay",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.UserWalletAccount(userID), Debit: amount, Credit: 0},
			{AccountCode: domainledger.ShipmentEscrowAccount(shipmentID), Debit: 0, Credit: amount},
		},
	})
	if err != nil {
		t.Fatalf("PostJournal escrow error = %v", err)
	}

	mid, err := getTreasury.Execute(ctx, "IRT")
	if err != nil {
		t.Fatalf("GetTreasurySummary() after escrow error = %v", err)
	}
	if delta := mid.Accounts["AGGREGATE_WALLET_LIABILITY"] - after.Accounts["AGGREGATE_WALLET_LIABILITY"]; delta != -amount {
		t.Fatalf("wallet aggregate delta = %d, want %d", delta, -amount)
	}
	if delta := mid.Accounts["AGGREGATE_ESCROW_LIABILITY"] - after.Accounts["AGGREGATE_ESCROW_LIABILITY"]; delta != amount {
		t.Fatalf("escrow aggregate delta = %d, want %d", delta, amount)
	}
}
