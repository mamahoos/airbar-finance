//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/mamahoos/airbar-finance/internal/domain/ledger"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
)

func TestLedgerRepositoryPostJournalIntegration(t *testing.T) {
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

	repo := NewLedgerRepository(pool)
	uc := ledgeruc.NewPostJournal(repo, nil)

	shipmentID := fmt.Sprintf("integration-sh-%s", uuid.NewString()[:8])
	userID := fmt.Sprintf("user-a-%s", uuid.NewString()[:8])
	refID := shipmentID + ":wallet-pay"

	journal, err := uc.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: ledger.RefTypeWalletToEscrow,
		RefID:   refID,
		Lines: []ledger.EntryLine{
			{AccountCode: ledger.UserWalletAccount(userID), Debit: 2500, Credit: 0},
			{AccountCode: ledger.ShipmentEscrowAccount(shipmentID), Debit: 0, Credit: 2500},
		},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if journal.ID == "" {
		t.Fatal("expected journal ID")
	}

	debit, credit, err := repo.SumByAccount(ctx, ledger.UserWalletAccount(userID))
	if err != nil {
		t.Fatalf("SumByAccount() error = %v", err)
	}
	if debit != 2500 || credit != 0 {
		t.Fatalf("user sums = (%d, %d), want (2500, 0)", debit, credit)
	}

	escrowDebit, escrowCredit, err := repo.SumByAccount(ctx, ledger.ShipmentEscrowAccount(shipmentID))
	if err != nil {
		t.Fatalf("SumByAccount() escrow error = %v", err)
	}
	if escrowDebit != 0 || escrowCredit != 2500 {
		t.Fatalf("escrow sums = (%d, %d), want (0, 2500)", escrowDebit, escrowCredit)
	}

	_, err = uc.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: ledger.RefTypeWalletToEscrow,
		RefID:   refID,
		Lines: []ledger.EntryLine{
			{AccountCode: ledger.UserWalletAccount(userID), Debit: 2500, Credit: 0},
			{AccountCode: ledger.ShipmentEscrowAccount(shipmentID), Debit: 0, Credit: 2500},
		},
	})
	if err != ledger.ErrDuplicateJournal {
		t.Fatalf("duplicate Execute() error = %v, want ErrDuplicateJournal", err)
	}
}
