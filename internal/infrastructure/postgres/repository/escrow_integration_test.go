//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	escrowuc "github.com/mamahoos/airbar-finance/internal/usecase/escrow"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
)

func TestEscrowWalletFlowIntegration(t *testing.T) {
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
	escrowRepo := NewEscrowRepository(pool)

	ensureWallet := walletuc.NewEnsureWalletAccount(walletRepo)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, ensureWallet)
	getBalance := walletuc.NewGetBalance(ledgerRepo)

	createEscrow := escrowuc.NewCreateEscrow(escrowRepo)
	payFromWallet := escrowuc.NewPayFromWallet(pool, escrowRepo, postJournal, getBalance)
	markDelivered := escrowuc.NewMarkDelivered(escrowRepo)
	releaseEscrow := escrowuc.NewReleaseEscrow(pool, escrowRepo, postJournal, ledgerRepo, 10)
	freezeEscrow := escrowuc.NewFreezeEscrow(escrowRepo)
	refundEscrow := escrowuc.NewRefundEscrow(pool, escrowRepo, postJournal, ledgerRepo)

	suffix := uuid.NewString()[:8]
	shipmentID := fmt.Sprintf("sh-%s", suffix)
	payerID := fmt.Sprintf("payer-%s", suffix)
	carrierID := fmt.Sprintf("carrier-%s", suffix)
	amount := int64(10000)

	escrow, err := createEscrow.Execute(ctx, escrowuc.CreateEscrowInput{
		ShipmentID:    shipmentID,
		CarrierUserID: carrierID,
		PayerUserID:   payerID,
		Amount:        amount,
	})
	if err != nil {
		t.Fatalf("CreateEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusCreated {
		t.Fatalf("status = %q, want CREATED", escrow.Status)
	}

	_, err = postJournal.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletTopup,
		RefID:   payerID + ":topup",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.AccountIRPSPClearing, Debit: amount, Credit: 0},
			{AccountCode: domainledger.UserWalletAccount(payerID), Debit: 0, Credit: amount},
		},
	})
	if err != nil {
		t.Fatalf("topup Execute() error = %v", err)
	}

	escrow, err = payFromWallet.Execute(ctx, escrowuc.PayFromWalletInput{
		ShipmentID:  shipmentID,
		PayerUserID: payerID,
		Amount:      amount,
	})
	if err != nil {
		t.Fatalf("PayFromWallet() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusFunded {
		t.Fatalf("status after pay = %q, want FUNDED", escrow.Status)
	}
	if escrow.FundingSource != domainescrow.FundingSourceWallet {
		t.Fatalf("funding_source = %q, want WALLET", escrow.FundingSource)
	}

	payerBalance, err := getBalance.Execute(ctx, payerID)
	if err != nil {
		t.Fatalf("payer GetBalance() error = %v", err)
	}
	if payerBalance != 0 {
		t.Fatalf("payer balance after pay = %d, want 0", payerBalance)
	}

	escrow, err = markDelivered.Execute(ctx, escrowuc.MarkDeliveredInput{ShipmentID: shipmentID})
	if err != nil {
		t.Fatalf("MarkDelivered() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusDisputeWindow {
		t.Fatalf("status after delivery = %q, want DISPUTE_WINDOW", escrow.Status)
	}

	feeDebitBefore, feeCreditBefore, err := ledgerRepo.SumByAccount(ctx, domainledger.AccountAirbarFeeRevenue)
	if err != nil {
		t.Fatalf("fee SumByAccount() before release error = %v", err)
	}

	escrow, err = releaseEscrow.Execute(ctx, escrowuc.ReleaseEscrowInput{ShipmentID: shipmentID})
	if err != nil {
		t.Fatalf("ReleaseEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusReleased {
		t.Fatalf("status after release = %q, want RELEASED", escrow.Status)
	}

	carrierBalance, err := getBalance.Execute(ctx, carrierID)
	if err != nil {
		t.Fatalf("carrier GetBalance() error = %v", err)
	}
	if carrierBalance != 9000 {
		t.Fatalf("carrier balance = %d, want 9000", carrierBalance)
	}

	feeDebit, feeCredit, err := ledgerRepo.SumByAccount(ctx, domainledger.AccountAirbarFeeRevenue)
	if err != nil {
		t.Fatalf("fee SumByAccount() error = %v", err)
	}
	feeDelta := (feeCredit - feeDebit) - (feeCreditBefore - feeDebitBefore)
	if feeDelta != 1000 {
		t.Fatalf("fee revenue delta = %d, want 1000", feeDelta)
	}

	// Gate F3 dispute path: freeze blocks release; refund credits payer wallet.
	shipmentID2 := fmt.Sprintf("sh2-%s", suffix)
	escrow2, err := createEscrow.Execute(ctx, escrowuc.CreateEscrowInput{
		ShipmentID:    shipmentID2,
		CarrierUserID: carrierID,
		PayerUserID:   payerID,
		Amount:        amount,
	})
	if err != nil {
		t.Fatalf("CreateEscrow(2) error = %v", err)
	}
	_ = escrow2

	_, err = postJournal.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletTopup,
		RefID:   payerID + ":topup2",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.AccountIRPSPClearing, Debit: amount, Credit: 0},
			{AccountCode: domainledger.UserWalletAccount(payerID), Debit: 0, Credit: amount},
		},
	})
	if err != nil {
		t.Fatalf("topup2 Execute() error = %v", err)
	}

	_, err = payFromWallet.Execute(ctx, escrowuc.PayFromWalletInput{
		ShipmentID:  shipmentID2,
		PayerUserID: payerID,
		Amount:      amount,
	})
	if err != nil {
		t.Fatalf("PayFromWallet(2) error = %v", err)
	}

	_, err = markDelivered.Execute(ctx, escrowuc.MarkDeliveredInput{ShipmentID: shipmentID2})
	if err != nil {
		t.Fatalf("MarkDelivered(2) error = %v", err)
	}

	_, err = freezeEscrow.Execute(ctx, escrowuc.FreezeEscrowInput{ShipmentID: shipmentID2})
	if err != nil {
		t.Fatalf("FreezeEscrow() error = %v", err)
	}

	_, err = releaseEscrow.Execute(ctx, escrowuc.ReleaseEscrowInput{ShipmentID: shipmentID2})
	if err == nil {
		t.Fatal("ReleaseEscrow from FROZEN should fail")
	}

	escrow2, err = refundEscrow.Execute(ctx, escrowuc.RefundEscrowInput{ShipmentID: shipmentID2})
	if err != nil {
		t.Fatalf("RefundEscrow() error = %v", err)
	}
	if escrow2.Status != domainescrow.StatusRefunded {
		t.Fatalf("status after refund = %q, want REFUNDED", escrow2.Status)
	}

	payerBalance, err = getBalance.Execute(ctx, payerID)
	if err != nil {
		t.Fatalf("payer GetBalance() after refund error = %v", err)
	}
	if payerBalance != amount {
		t.Fatalf("payer balance after refund = %d, want %d", payerBalance, amount)
	}
}
