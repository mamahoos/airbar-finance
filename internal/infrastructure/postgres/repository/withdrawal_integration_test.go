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
	createWithdrawal := withdrawaluc.NewCreateWithdrawal(pool, withdrawalRepo, postJournal, getBalance, nil)
	approveWithdrawal := withdrawaluc.NewApproveWithdrawal(withdrawalRepo, nil)
	markWithdrawalSent := withdrawaluc.NewMarkWithdrawalSent(withdrawalRepo, nil)
	settleWithdrawal := withdrawaluc.NewSettleWithdrawal(withdrawalRepo, nil)
	failWithdrawal := withdrawaluc.NewFailWithdrawal(pool, withdrawalRepo, postJournal, nil)
	rejectWithdrawal := withdrawaluc.NewRejectWithdrawal(pool, withdrawalRepo, postJournal, nil)

	processUser := fmt.Sprintf("wd-process-%s", uuid.NewString()[:8])
	rejectUser := fmt.Sprintf("wd-reject-%s", uuid.NewString()[:8])
	failUser := fmt.Sprintf("wd-fail-%s", uuid.NewString()[:8])
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
	fundWallet(failUser, failUser+":topup", amount)

	payoutBeforeDebit, payoutBeforeCredit, err := ledgerRepo.SumByAccount(ctx, domainledger.AccountIRPayoutClearing)
	if err != nil {
		t.Fatalf("payout SumByAccount() before reserve error = %v", err)
	}

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
	payoutDelta := (payoutCredit - payoutDebit) - (payoutBeforeCredit - payoutBeforeDebit)
	if payoutDelta != amount {
		t.Fatalf("payout clearing delta = %d, want %d", payoutDelta, amount)
	}

	withdrawal, err = approveWithdrawal.Execute(ctx, withdrawal.ID)
	if err != nil {
		t.Fatalf("ApproveWithdrawal() error = %v", err)
	}
	if withdrawal.Status != domainwithdrawal.StatusApproved {
		t.Fatalf("status after approve = %q, want APPROVED", withdrawal.Status)
	}

	withdrawal, err = markWithdrawalSent.Execute(ctx, withdrawaluc.MarkWithdrawalSentInput{
		WithdrawalID:  withdrawal.ID,
		ProviderRef:   "bank-ref-1",
		PayoutChannel: "PAYA",
		ReceiptURL:    "https://receipts.example/bank-ref-1",
	})
	if err != nil {
		t.Fatalf("MarkWithdrawalSent() error = %v", err)
	}
	if withdrawal.Status != domainwithdrawal.StatusSentToBank {
		t.Fatalf("status after sent = %q, want SENT_TO_BANK", withdrawal.Status)
	}
	if withdrawal.ProviderRef != "bank-ref-1" || withdrawal.PayoutChannel != "PAYA" || withdrawal.ReceiptURL == "" {
		t.Fatalf("receipt fields not persisted: %+v", withdrawal)
	}
	withdrawal, err = settleWithdrawal.Execute(ctx, withdrawal.ID)
	if err != nil {
		t.Fatalf("SettleWithdrawal() error = %v", err)
	}
	if withdrawal.Status != domainwithdrawal.StatusSettled {
		t.Fatalf("status after settle = %q, want SETTLED", withdrawal.Status)
	}
	if withdrawal.ProcessedAt == nil {
		t.Fatal("settled withdrawal should set ProcessedAt")
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

	failWD, err := createWithdrawal.Execute(ctx, withdrawaluc.CreateWithdrawalInput{
		UserID:               failUser,
		Amount:               amount,
		DestinationIBAN:      "IR120000000000000000000003",
		UserActive:           true,
		FinancialKycApproved: true,
	})
	if err != nil {
		t.Fatalf("CreateWithdrawal(fail) error = %v", err)
	}
	failWD, err = approveWithdrawal.Execute(ctx, failWD.ID)
	if err != nil {
		t.Fatalf("ApproveWithdrawal(fail) error = %v", err)
	}
	failWD, err = failWithdrawal.Execute(ctx, withdrawaluc.FailWithdrawalInput{
		WithdrawalID: failWD.ID,
		Reason:       "bank rejected payout",
	})
	if err != nil {
		t.Fatalf("FailWithdrawal() error = %v", err)
	}
	if failWD.Status != domainwithdrawal.StatusFailed {
		t.Fatalf("status after fail = %q, want FAILED", failWD.Status)
	}
	balance, err = getBalance.Execute(ctx, failUser)
	if err != nil {
		t.Fatalf("GetBalance() after fail error = %v", err)
	}
	if balance != amount {
		t.Fatalf("wallet balance after fail = %d, want %d", balance, amount)
	}
}
