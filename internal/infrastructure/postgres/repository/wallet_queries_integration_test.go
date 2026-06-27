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

func TestWalletQueriesIntegration(t *testing.T) {
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
	postJournalUC := ledgeruc.NewPostJournal(ledgerRepo, ensureUC)
	getBalanceUC := walletuc.NewGetBalance(ledgerRepo)
	getWalletUC := walletuc.NewGetWallet(getBalanceUC)
	listTxUC := walletuc.NewListWalletTransactions(ledgerRepo)

	userID := fmt.Sprintf("wallet-query-%s", uuid.NewString()[:8])
	refTopup := userID + ":topup"
	refPay := userID + ":pay"

	_, err = postJournalUC.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletTopup,
		RefID:   refTopup,
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.AccountIRPSPClearing, Debit: 20000, Credit: 0},
			{AccountCode: domainledger.UserWalletAccount(userID), Debit: 0, Credit: 20000},
		},
	})
	if err != nil {
		t.Fatalf("topup Execute() error = %v", err)
	}

	_, err = postJournalUC.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletToEscrow,
		RefID:   refPay,
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.UserWalletAccount(userID), Debit: 5000, Credit: 0},
			{AccountCode: domainledger.ShipmentEscrowAccount("sh-" + userID), Debit: 0, Credit: 5000},
		},
	})
	if err != nil {
		t.Fatalf("pay Execute() error = %v", err)
	}

	wallet, err := getWalletUC.Execute(ctx, userID, "IRT")
	if err != nil {
		t.Fatalf("GetWallet() error = %v", err)
	}
	if wallet.Balance != 15000 {
		t.Fatalf("wallet balance = %d, want 15000", wallet.Balance)
	}
	if wallet.AccountCode != domainledger.UserWalletAccount(userID).String() {
		t.Fatalf("account_code = %q", wallet.AccountCode)
	}

	items, err := listTxUC.Execute(ctx, userID, "IRT")
	if err != nil {
		t.Fatalf("ListWalletTransactions() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("transaction count = %d, want 2", len(items))
	}
	if items[0].RefType != string(domainledger.RefTypeWalletToEscrow) {
		t.Fatalf("latest ref_type = %q, want WALLET_TO_ESCROW", items[0].RefType)
	}
	if items[1].RefType != string(domainledger.RefTypeWalletTopup) {
		t.Fatalf("older ref_type = %q, want WALLET_TOPUP", items[1].RefType)
	}
}
