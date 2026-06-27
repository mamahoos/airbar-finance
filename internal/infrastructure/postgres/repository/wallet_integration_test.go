//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
)

func TestWalletEnsureAndGetBalanceIntegration(t *testing.T) {
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
	walletRepo := NewWalletRepository(pool)
	ensureUC := walletuc.NewEnsureWalletAccount(walletRepo)
	getBalanceUC := walletuc.NewGetBalance(ledgerRepo)
	postJournalUC := ledgeruc.NewPostJournal(ledgerRepo, ensureUC)

	userID := fmt.Sprintf("wallet-user-%s", uuid.NewString()[:8])
	shipmentID := fmt.Sprintf("wallet-sh-%s", uuid.NewString()[:8])
	refTopup := userID + ":topup"
	refPay := shipmentID + ":wallet-pay"

	_, err = postJournalUC.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletTopup,
		RefID:   refTopup,
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.AccountIRPSPClearing, Debit: 10000, Credit: 0},
			{AccountCode: domainledger.UserWalletAccount(userID), Debit: 0, Credit: 10000},
		},
	})
	if err != nil {
		t.Fatalf("topup Execute() error = %v", err)
	}

	account, err := walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if account == nil {
		t.Fatal("expected wallet account to be lazily created")
	}
	if account.AccountCode != domainledger.UserWalletAccount(userID).String() {
		t.Fatalf("account_code = %q", account.AccountCode)
	}

	balance, err := getBalanceUC.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if balance != 10000 {
		t.Fatalf("balance after topup = %d, want 10000", balance)
	}

	_, err = postJournalUC.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletToEscrow,
		RefID:   refPay,
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.UserWalletAccount(userID), Debit: 3000, Credit: 0},
			{AccountCode: domainledger.ShipmentEscrowAccount(shipmentID), Debit: 0, Credit: 3000},
		},
	})
	if err != nil {
		t.Fatalf("pay Execute() error = %v", err)
	}

	balance, err = getBalanceUC.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("GetBalance() after pay error = %v", err)
	}
	if balance != 7000 {
		t.Fatalf("balance after pay = %d, want 7000", balance)
	}
}
