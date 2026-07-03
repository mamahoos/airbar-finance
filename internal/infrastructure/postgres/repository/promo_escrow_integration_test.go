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
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
	credituc "github.com/mamahoos/airbar-finance/internal/usecase/credit"
	escrowuc "github.com/mamahoos/airbar-finance/internal/usecase/escrow"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
)

func TestEscrowPromoCreditPayAndRefundIntegration(t *testing.T) {
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
	creditRepo := NewCreditRepository(pool)
	escrowRepo := NewEscrowRepository(pool)
	ensureWallet := walletuc.NewEnsureWalletAccount(walletRepo)
	ensureCredit := credituc.NewEnsureCreditAccount(creditRepo)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, ensureWallet)
	getWalletBalance := walletuc.NewGetBalance(ledgerRepo)
	getCreditBalance := credituc.NewGetBalance(ledgerRepo)
	auditEmitter := audituc.NewEmitter(NewFinanceEventRepository(pool))

	grantCredit := credituc.NewGrantCredit(pool, creditRepo, ensureCredit, postJournal, auditEmitter)
	createEscrow := escrowuc.NewCreateEscrow(escrowRepo, nil)
	payFromWallet := escrowuc.NewPayFromWallet(pool, escrowRepo, postJournal, getWalletBalance, getCreditBalance, ensureCredit, nil)
	refundEscrow := escrowuc.NewRefundEscrow(pool, escrowRepo, postJournal, ledgerRepo, ensureCredit, nil)

	suffix := uuid.NewString()[:8]
	shipmentID := fmt.Sprintf("promo-sh-%s", suffix)
	payerID := fmt.Sprintf("promo-payer-%s", suffix)
	carrierID := fmt.Sprintf("promo-carrier-%s", suffix)
	amount := int64(10000)
	promoGrant := int64(6000)

	_, err = grantCredit.Execute(ctx, credituc.GrantCreditInput{
		UserID:         payerID,
		Amount:         promoGrant,
		Reason:         "welcome bonus",
		GrantedBy:      "admin-test",
		IdempotencyKey: uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("GrantCredit() error = %v", err)
	}

	_, err = postJournal.Execute(ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletTopup,
		RefID:   payerID + ":topup",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.AccountIRPSPClearing, Debit: 5000, Credit: 0},
			{AccountCode: domainledger.UserWalletAccount(payerID), Debit: 0, Credit: 5000},
		},
	})
	if err != nil {
		t.Fatalf("wallet topup error = %v", err)
	}

	_, err = createEscrow.Execute(ctx, escrowuc.CreateEscrowInput{
		ShipmentID:    shipmentID,
		CarrierUserID: carrierID,
		PayerUserID:   payerID,
		Amount:        amount,
	})
	if err != nil {
		t.Fatalf("CreateEscrow() error = %v", err)
	}

	escrow, err := payFromWallet.Execute(ctx, escrowuc.PayFromWalletInput{
		ShipmentID:  shipmentID,
		PayerUserID: payerID,
		Amount:      amount,
	})
	if err != nil {
		t.Fatalf("PayFromWallet() error = %v", err)
	}
	if escrow.FundingSource != domainescrow.FundingSourceMixed {
		t.Fatalf("funding_source = %q, want MIXED", escrow.FundingSource)
	}
	if escrow.PromoCreditFunded != 6000 {
		t.Fatalf("promo_credit_funded = %d, want 6000", escrow.PromoCreditFunded)
	}

	promoBalance, err := getCreditBalance.Execute(ctx, payerID)
	if err != nil {
		t.Fatalf("promo balance error = %v", err)
	}
	if promoBalance != 0 {
		t.Fatalf("promo balance after pay = %d, want 0", promoBalance)
	}

	walletBalance, err := getWalletBalance.Execute(ctx, payerID)
	if err != nil {
		t.Fatalf("wallet balance error = %v", err)
	}
	if walletBalance != 1000 {
		t.Fatalf("wallet balance after pay = %d, want 1000", walletBalance)
	}

	escrow, err = refundEscrow.Execute(ctx, escrowuc.RefundEscrowInput{ShipmentID: shipmentID})
	if err != nil {
		t.Fatalf("RefundEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusRefunded {
		t.Fatalf("status after refund = %q, want REFUNDED", escrow.Status)
	}
	if escrow.PromoCreditFunded != 0 {
		t.Fatalf("promo_credit_funded after refund = %d, want 0", escrow.PromoCreditFunded)
	}

	promoBalance, err = getCreditBalance.Execute(ctx, payerID)
	if err != nil {
		t.Fatalf("promo balance after refund error = %v", err)
	}
	if promoBalance != 6000 {
		t.Fatalf("promo balance after refund = %d, want 6000", promoBalance)
	}

	walletBalance, err = getWalletBalance.Execute(ctx, payerID)
	if err != nil {
		t.Fatalf("wallet balance after refund error = %v", err)
	}
	if walletBalance != 5000 {
		t.Fatalf("wallet balance after refund = %d, want 5000", walletBalance)
	}
}
