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
	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/zibal"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
	credituc "github.com/mamahoos/airbar-finance/internal/usecase/credit"
	escrowuc "github.com/mamahoos/airbar-finance/internal/usecase/escrow"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	paymentuc "github.com/mamahoos/airbar-finance/internal/usecase/payment"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
)

type escrowLifecycleHarness struct {
	ctx context.Context

	createEscrow         *escrowuc.CreateEscrow
	fundEscrow           *escrowuc.FundEscrow
	payFromWallet        *escrowuc.PayFromWallet
	markDelivered        *escrowuc.MarkDelivered
	releaseEscrow        *escrowuc.ReleaseEscrow
	partialRefundEscrow  *escrowuc.PartialRefundEscrow
	grantCredit          *credituc.GrantCredit
	createPaymentOrder   *paymentuc.CreatePaymentOrder
	verifyPaymentOrder   *paymentuc.VerifyPaymentOrder
	getEscrow            *escrowuc.GetEscrow
	postJournal          *ledgeruc.PostJournal
	getWalletBalance     *walletuc.GetBalance
	getCreditBalance     *credituc.GetBalance
	ledgerRepo           *LedgerRepository
}

func newEscrowLifecycleHarness(t *testing.T) *escrowLifecycleHarness {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	t.Cleanup(pool.Close)

	mockZibal := zibal.NewMockServer()
	t.Cleanup(mockZibal.Close)

	ledgerRepo := NewLedgerRepository(pool)
	walletRepo := NewWalletRepository(pool)
	creditRepo := NewCreditRepository(pool)
	escrowRepo := NewEscrowRepository(pool)
	paymentRepo := NewPaymentRepository(pool)
	providerRepo := NewProviderEventRepository(pool)
	zibalClient := mockZibal.Client()

	ensureWallet := walletuc.NewEnsureWalletAccount(walletRepo)
	ensureCredit := credituc.NewEnsureCreditAccount(creditRepo)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, ensureWallet)
	getWalletBalance := walletuc.NewGetBalance(ledgerRepo)
	getCreditBalance := credituc.NewGetBalance(ledgerRepo)
	auditEmitter := audituc.NewEmitter(NewFinanceEventRepository(pool))

	fundEscrow := escrowuc.NewFundEscrow(pool, escrowRepo, postJournal, nil)
	verifyOrder := paymentuc.NewVerifyOrder(pool, paymentRepo, providerRepo, zibalClient, fundEscrow, postJournal, nil)

	return &escrowLifecycleHarness{
		ctx:                 ctx,
		createEscrow:        escrowuc.NewCreateEscrow(escrowRepo, nil),
		fundEscrow:          fundEscrow,
		payFromWallet:       escrowuc.NewPayFromWallet(pool, escrowRepo, postJournal, getWalletBalance, getCreditBalance, ensureCredit, nil),
		markDelivered:       escrowuc.NewMarkDelivered(escrowRepo, nil),
		releaseEscrow:       escrowuc.NewReleaseEscrow(pool, escrowRepo, postJournal, ledgerRepo, 10, nil),
		partialRefundEscrow: escrowuc.NewPartialRefundEscrow(pool, escrowRepo, postJournal, ledgerRepo, ensureCredit, auditEmitter),
		grantCredit:         credituc.NewGrantCredit(pool, creditRepo, ensureCredit, postJournal, auditEmitter),
		createPaymentOrder: paymentuc.NewCreatePaymentOrder(
			paymentRepo,
			escrowRepo,
			providerRepo,
			zibalClient,
			"http://localhost:8080",
			nil,
		),
		verifyPaymentOrder: paymentuc.NewVerifyPaymentOrder(verifyOrder),
		getEscrow:          escrowuc.NewGetEscrow(escrowRepo),
		postJournal:        postJournal,
		getWalletBalance:   getWalletBalance,
		getCreditBalance:   getCreditBalance,
		ledgerRepo:         ledgerRepo,
	}
}

func TestEscrowLifecyclePSPReleaseIntegration(t *testing.T) {
	h := newEscrowLifecycleHarness(t)
	suffix := uuid.NewString()[:8]
	shipmentID := fmt.Sprintf("life-psp-%s", suffix)
	payerID := fmt.Sprintf("payer-%s", suffix)
	carrierID := fmt.Sprintf("carrier-%s", suffix)
	amount := int64(30000)

	_, err := h.createEscrow.Execute(h.ctx, escrowuc.CreateEscrowInput{
		ShipmentID:    shipmentID,
		CarrierUserID: carrierID,
		PayerUserID:   payerID,
		Amount:        amount,
	})
	if err != nil {
		t.Fatalf("CreateEscrow() error = %v", err)
	}

	order, err := h.createPaymentOrder.Execute(h.ctx, paymentuc.CreatePaymentOrderInput{
		ShipmentID:  shipmentID,
		PayerUserID: payerID,
		Amount:      amount,
		SuccessURL:  "https://app/success",
		FailureURL:  "https://app/failure",
		Description: "lifecycle psp",
		AgreedPrice: amount,
	})
	if err != nil {
		t.Fatalf("CreatePaymentOrder() error = %v", err)
	}

	order, err = h.verifyPaymentOrder.Execute(h.ctx, order.ID, order.Authority)
	if err != nil {
		t.Fatalf("VerifyPaymentOrder() error = %v", err)
	}
	if order.Status != domainpayment.StatusConfirmed {
		t.Fatalf("order status = %q, want CONFIRMED", order.Status)
	}

	escrow, err := h.getEscrow.Execute(h.ctx, shipmentID)
	if err != nil {
		t.Fatalf("GetEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusFunded {
		t.Fatalf("escrow status = %q, want FUNDED", escrow.Status)
	}
	if escrow.FundingSource != domainescrow.FundingSourcePSP {
		t.Fatalf("funding_source = %q, want PSP", escrow.FundingSource)
	}

	escrow, err = h.markDelivered.Execute(h.ctx, escrowuc.MarkDeliveredInput{ShipmentID: shipmentID})
	if err != nil {
		t.Fatalf("MarkDelivered() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusDisputeWindow {
		t.Fatalf("status after delivery = %q, want DISPUTE_WINDOW", escrow.Status)
	}

	escrow, err = h.releaseEscrow.Execute(h.ctx, escrowuc.ReleaseEscrowInput{ShipmentID: shipmentID})
	if err != nil {
		t.Fatalf("ReleaseEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusReleased {
		t.Fatalf("status after release = %q, want RELEASED", escrow.Status)
	}

	carrierBalance, err := h.getWalletBalance.Execute(h.ctx, carrierID)
	if err != nil {
		t.Fatalf("carrier balance error = %v", err)
	}
	if carrierBalance != 27000 {
		t.Fatalf("carrier balance = %d, want 27000", carrierBalance)
	}
}

func TestEscrowLifecyclePartialRefundThenReleaseIntegration(t *testing.T) {
	h := newEscrowLifecycleHarness(t)
	suffix := uuid.NewString()[:8]
	shipmentID := fmt.Sprintf("life-partial-%s", suffix)
	payerID := fmt.Sprintf("payer-%s", suffix)
	carrierID := fmt.Sprintf("carrier-%s", suffix)
	amount := int64(20000)
	partialRefund := int64(7000)

	_, err := h.createEscrow.Execute(h.ctx, escrowuc.CreateEscrowInput{
		ShipmentID:    shipmentID,
		CarrierUserID: carrierID,
		PayerUserID:   payerID,
		Amount:        amount,
	})
	if err != nil {
		t.Fatalf("CreateEscrow() error = %v", err)
	}

	_, err = h.postJournal.Execute(h.ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletTopup,
		RefID:   payerID + ":topup",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.AccountIRPSPClearing, Debit: amount, Credit: 0},
			{AccountCode: domainledger.UserWalletAccount(payerID), Debit: 0, Credit: amount},
		},
	})
	if err != nil {
		t.Fatalf("topup error = %v", err)
	}

	escrow, err := h.payFromWallet.Execute(h.ctx, escrowuc.PayFromWalletInput{
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

	escrow, err = h.markDelivered.Execute(h.ctx, escrowuc.MarkDeliveredInput{ShipmentID: shipmentID})
	if err != nil {
		t.Fatalf("MarkDelivered() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusDisputeWindow {
		t.Fatalf("status after delivery = %q, want DISPUTE_WINDOW", escrow.Status)
	}

	escrow, err = h.partialRefundEscrow.Execute(h.ctx, escrowuc.PartialRefundEscrowInput{
		ShipmentID:   shipmentID,
		RefundAmount: partialRefund,
	})
	if err != nil {
		t.Fatalf("PartialRefundEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusPartiallyRefunded {
		t.Fatalf("status after partial refund = %q, want PARTIALLY_REFUNDED", escrow.Status)
	}

	remaining, err := escrowuc.EscrowBalance(h.ctx, h.ledgerRepo, shipmentID)
	if err != nil {
		t.Fatalf("EscrowBalance() error = %v", err)
	}
	if remaining != amount-partialRefund {
		t.Fatalf("remaining escrow balance = %d, want %d", remaining, amount-partialRefund)
	}

	payerBalance, err := h.getWalletBalance.Execute(h.ctx, payerID)
	if err != nil {
		t.Fatalf("payer balance error = %v", err)
	}
	if payerBalance != partialRefund {
		t.Fatalf("payer balance after partial refund = %d, want %d", payerBalance, partialRefund)
	}

	escrow, err = h.releaseEscrow.Execute(h.ctx, escrowuc.ReleaseEscrowInput{ShipmentID: shipmentID})
	if err != nil {
		t.Fatalf("ReleaseEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusReleased {
		t.Fatalf("status after release = %q, want RELEASED", escrow.Status)
	}

	carrierBalance, err := h.getWalletBalance.Execute(h.ctx, carrierID)
	if err != nil {
		t.Fatalf("carrier balance error = %v", err)
	}
	if carrierBalance != 11700 {
		t.Fatalf("carrier balance = %d, want 11700 (90%% of 13000 remainder)", carrierBalance)
	}
}

func TestEscrowLifecyclePromoThroughDisputeWindowIntegration(t *testing.T) {
	h := newEscrowLifecycleHarness(t)
	suffix := uuid.NewString()[:8]
	shipmentID := fmt.Sprintf("life-promo-%s", suffix)
	payerID := fmt.Sprintf("payer-%s", suffix)
	carrierID := fmt.Sprintf("carrier-%s", suffix)
	amount := int64(10000)
	promoGrant := int64(6000)
	partialRefund := int64(3000)

	_, err := h.grantCredit.Execute(h.ctx, credituc.GrantCreditInput{
		UserID:         payerID,
		Amount:         promoGrant,
		Reason:         "campaign bonus",
		GrantedBy:      "admin-test",
		IdempotencyKey: uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("GrantCredit() error = %v", err)
	}

	_, err = h.postJournal.Execute(h.ctx, ledgeruc.PostJournalInput{
		RefType: domainledger.RefTypeWalletTopup,
		RefID:   payerID + ":topup",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.AccountIRPSPClearing, Debit: 5000, Credit: 0},
			{AccountCode: domainledger.UserWalletAccount(payerID), Debit: 0, Credit: 5000},
		},
	})
	if err != nil {
		t.Fatalf("topup error = %v", err)
	}

	_, err = h.createEscrow.Execute(h.ctx, escrowuc.CreateEscrowInput{
		ShipmentID:    shipmentID,
		CarrierUserID: carrierID,
		PayerUserID:   payerID,
		Amount:        amount,
	})
	if err != nil {
		t.Fatalf("CreateEscrow() error = %v", err)
	}

	escrow, err := h.payFromWallet.Execute(h.ctx, escrowuc.PayFromWalletInput{
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

	escrow, err = h.markDelivered.Execute(h.ctx, escrowuc.MarkDeliveredInput{ShipmentID: shipmentID})
	if err != nil {
		t.Fatalf("MarkDelivered() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusDisputeWindow {
		t.Fatalf("status after delivery = %q, want DISPUTE_WINDOW", escrow.Status)
	}

	escrow, err = h.partialRefundEscrow.Execute(h.ctx, escrowuc.PartialRefundEscrowInput{
		ShipmentID:   shipmentID,
		RefundAmount: partialRefund,
	})
	if err != nil {
		t.Fatalf("PartialRefundEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusPartiallyRefunded {
		t.Fatalf("status after partial refund = %q, want PARTIALLY_REFUNDED", escrow.Status)
	}

	escrow, err = h.releaseEscrow.Execute(h.ctx, escrowuc.ReleaseEscrowInput{ShipmentID: shipmentID})
	if err != nil {
		t.Fatalf("ReleaseEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusReleased {
		t.Fatalf("status after release = %q, want RELEASED", escrow.Status)
	}

	carrierBalance, err := h.getWalletBalance.Execute(h.ctx, carrierID)
	if err != nil {
		t.Fatalf("carrier balance error = %v", err)
	}
	if carrierBalance != 6300 {
		t.Fatalf("carrier balance = %d, want 6300", carrierBalance)
	}
}
