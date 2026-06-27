//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
	withdrawaluc "github.com/mamahoos/airbar-finance/internal/usecase/withdrawal"
)

func TestWithdrawalReserveProcessRejectIntegration(t *testing.T) {
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
	withdrawalRepo := NewWithdrawalRepository(pool)

	ensureWallet := walletuc.NewEnsureWalletAccount(walletRepo)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, ensureWallet)
	getBalance := walletuc.NewGetBalance(ledgerRepo)
	createWithdrawal := withdrawaluc.NewCreateWithdrawal(pool, withdrawalRepo, postJournal, getBalance)
	processWithdrawal := withdrawaluc.NewProcessWithdrawal(withdrawalRepo)
	rejectWithdrawal := withdrawaluc.NewRejectWithdrawal(pool, withdrawalRepo, postJournal)

	processUser := fmt.Sprintf("wd-process-%s", uuid.NewString()[:8])
	rejectUser := fmt.Sprintf("wd-reject-%s", uuid.NewString()[:8])
	amount := int64(8000)

	fundWallet := func(userID, refID string, fundAmount int64) {
		_, err := postJournal.Execute(ctx, ledgeruc.PostJournalInput{
			RefType: domainledger.RefTypeWalletTopup,
			RefID:   refID,
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.AccountIRPSPClearing, Debit: fundAmount, Credit: 0},
				{AccountCode: domainledger.UserWalletAccount(userID), Debit: 0, Credit: fundAmount},
			},
		})
		if err != nil {
			t.Fatalf("fund wallet %s error = %v", userID, err)
		}
	}

	fundWallet(processUser, processUser+":topup", amount)
	fundWallet(rejectUser, rejectUser+":topup", amount)

	withdrawal, err := createWithdrawal.Execute(ctx, withdrawaluc.CreateWithdrawalInput{
		UserID:               processUser,
		Amount:               amount,
		DestinationIBAN:      "IR120000000000000000000001",
		UserActive:           true,
		FinancialKycApproved: true,
	})
	if err != nil {
		t.Fatalf("CreateWithdrawal(process) error = %v", err)
	}
	if withdrawal.Status != domainwithdrawal.StatusPending {
		t.Fatalf("status = %q, want PENDING", withdrawal.Status)
	}
	if withdrawal.DestinationHash != domainwithdrawal.HashDestination("IR120000000000000000000001") {
		t.Fatal("expected hashed destination, not plain IBAN")
	}

	balance, err := getBalance.Execute(ctx, processUser)
	if err != nil {
		t.Fatalf("GetBalance() after reserve error = %v", err)
	}
	if balance != 0 {
		t.Fatalf("wallet balance after reserve = %d, want 0", balance)
	}

	payoutDebit, payoutCredit, err := ledgerRepo.SumByAccount(ctx, domainledger.AccountIRPayoutClearing)
	if err != nil {
		t.Fatalf("payout SumByAccount() error = %v", err)
	}
	if payoutCredit-payoutDebit != amount {
		t.Fatalf("payout clearing = %d, want %d", payoutCredit-payoutDebit, amount)
	}

	withdrawal, err = processWithdrawal.Execute(ctx, withdrawaluc.ProcessWithdrawalInput{
		WithdrawalID: withdrawal.ID,
		ProviderRef:  "bank-ref-1",
	})
	if err != nil {
		t.Fatalf("ProcessWithdrawal() error = %v", err)
	}
	if withdrawal.Status != domainwithdrawal.StatusCompleted {
		t.Fatalf("status after process = %q, want COMPLETED", withdrawal.Status)
	}

	rejectWD, err := createWithdrawal.Execute(ctx, withdrawaluc.CreateWithdrawalInput{
		UserID:               rejectUser,
		Amount:               amount,
		DestinationIBAN:      "IR120000000000000000000002",
		UserActive:           true,
		FinancialKycApproved: true,
	})
	if err != nil {
		t.Fatalf("CreateWithdrawal(reject) error = %v", err)
	}

	rejectWD, err = rejectWithdrawal.Execute(ctx, withdrawaluc.RejectWithdrawalInput{
		WithdrawalID: rejectWD.ID,
		Reason:       "admin reject test",
	})
	if err != nil {
		t.Fatalf("RejectWithdrawal() error = %v", err)
	}
	if rejectWD.Status != domainwithdrawal.StatusRejected {
		t.Fatalf("status after reject = %q, want REJECTED", rejectWD.Status)
	}

	balance, err = getBalance.Execute(ctx, rejectUser)
	if err != nil {
		t.Fatalf("GetBalance() after reject error = %v", err)
	}
	if balance != amount {
		t.Fatalf("wallet balance after reject = %d, want %d", balance, amount)
	}
}
