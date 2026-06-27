//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/zibal"
	escrowuc "github.com/mamahoos/airbar-finance/internal/usecase/escrow"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	paymentuc "github.com/mamahoos/airbar-finance/internal/usecase/payment"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
)

func TestPaymentDirectFlowIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	mockZibal := zibal.NewMockServer()
	defer mockZibal.Close()

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close()

	ledgerRepo := NewLedgerRepository(pool)
	walletRepo := NewWalletRepository(pool)
	escrowRepo := NewEscrowRepository(pool)
	paymentRepo := NewPaymentRepository(pool)
	providerRepo := NewProviderEventRepository(pool)
	zibalClient := mockZibal.Client()

	ensureWallet := walletuc.NewEnsureWalletAccount(walletRepo)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, ensureWallet)

	createEscrow := escrowuc.NewCreateEscrow(escrowRepo, nil)
	fundEscrow := escrowuc.NewFundEscrow(pool, escrowRepo, postJournal, nil)
	verifyOrder := paymentuc.NewVerifyOrder(pool, paymentRepo, providerRepo, zibalClient, fundEscrow, postJournal, nil)
	createPaymentOrder := paymentuc.NewCreatePaymentOrder(
		paymentRepo,
		escrowRepo,
		providerRepo,
		zibalClient,
		"http://localhost:8080",
		nil,
	)
	verifyPaymentOrder := paymentuc.NewVerifyPaymentOrder(verifyOrder)
	getEscrow := escrowuc.NewGetEscrow(escrowRepo)

	suffix := uuid.NewString()[:8]
	shipmentID := fmt.Sprintf("pay-sh-%s", suffix)
	payerID := fmt.Sprintf("payer-%s", suffix)
	carrierID := fmt.Sprintf("carrier-%s", suffix)
	amount := int64(25000)

	_, err = createEscrow.Execute(ctx, escrowuc.CreateEscrowInput{
		ShipmentID:    shipmentID,
		CarrierUserID: carrierID,
		PayerUserID:   payerID,
		Amount:        amount,
	})
	if err != nil {
		t.Fatalf("CreateEscrow() error = %v", err)
	}

	order, err := createPaymentOrder.Execute(ctx, paymentuc.CreatePaymentOrderInput{
		ShipmentID:  shipmentID,
		PayerUserID: payerID,
		Amount:      amount,
		SuccessURL:  "https://app/success",
		FailureURL:  "https://app/failure",
		Description: "shipment pay",
		AgreedPrice: amount,
	})
	if err != nil {
		t.Fatalf("CreatePaymentOrder() error = %v", err)
	}
	if order.Status != domainpayment.StatusPending {
		t.Fatalf("order status = %q, want PENDING", order.Status)
	}
	if order.Authority == "" || order.RedirectURL == "" {
		t.Fatal("expected authority and redirect_url")
	}

	order, err = verifyPaymentOrder.Execute(ctx, order.ID, order.Authority)
	if err != nil {
		t.Fatalf("VerifyPaymentOrder() error = %v", err)
	}
	if order.Status != domainpayment.StatusConfirmed {
		t.Fatalf("order status after verify = %q, want CONFIRMED", order.Status)
	}

	escrow, err := getEscrow.Execute(ctx, shipmentID)
	if err != nil {
		t.Fatalf("GetEscrow() error = %v", err)
	}
	if escrow.Status != domainescrow.StatusFunded {
		t.Fatalf("escrow status = %q, want FUNDED", escrow.Status)
	}
	if escrow.FundingSource != domainescrow.FundingSourcePSP {
		t.Fatalf("funding_source = %q, want PSP", escrow.FundingSource)
	}

	balance, err := escrowuc.EscrowBalance(ctx, ledgerRepo, shipmentID)
	if err != nil {
		t.Fatalf("EscrowBalance() error = %v", err)
	}
	if balance != amount {
		t.Fatalf("escrow balance = %d, want %d", balance, amount)
	}

	providerCount, err := providerRepo.CountByPaymentOrderID(ctx, order.ID)
	if err != nil {
		t.Fatalf("CountByPaymentOrderID() error = %v", err)
	}
	if providerCount < 2 {
		t.Fatalf("provider events = %d, want at least REQUEST+VERIFY", providerCount)
	}
}

func TestWalletTopupFlowIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	mockZibal := zibal.NewMockServer()
	defer mockZibal.Close()

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close()

	ledgerRepo := NewLedgerRepository(pool)
	walletRepo := NewWalletRepository(pool)
	escrowRepo := NewEscrowRepository(pool)
	paymentRepo := NewPaymentRepository(pool)
	providerRepo := NewProviderEventRepository(pool)
	zibalClient := mockZibal.Client()

	ensureWallet := walletuc.NewEnsureWalletAccount(walletRepo)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, ensureWallet)
	fundEscrow := escrowuc.NewFundEscrow(pool, escrowRepo, postJournal, nil)
	verifyOrder := paymentuc.NewVerifyOrder(pool, paymentRepo, providerRepo, zibalClient, fundEscrow, postJournal, nil)
	createTopup := paymentuc.NewCreateWalletTopupOrder(paymentRepo, providerRepo, zibalClient, "http://localhost:8080", nil)
	verifyTopup := paymentuc.NewVerifyWalletTopupOrder(verifyOrder)
	getBalance := walletuc.NewGetBalance(ledgerRepo)

	userID := fmt.Sprintf("topup-user-%s", uuid.NewString()[:8])
	amount := int64(50000)

	order, err := createTopup.Execute(ctx, paymentuc.CreateWalletTopupOrderInput{
		UserID:      userID,
		Amount:      amount,
		SuccessURL:  "https://app/wallet/success",
		FailureURL:  "https://app/wallet/failure",
		Description: "wallet topup",
	})
	if err != nil {
		t.Fatalf("CreateWalletTopupOrder() error = %v", err)
	}

	order, err = verifyTopup.Execute(ctx, order.ID, order.Authority)
	if err != nil {
		t.Fatalf("VerifyWalletTopupOrder() error = %v", err)
	}
	if order.Status != domainpayment.StatusConfirmed {
		t.Fatalf("order status = %q, want CONFIRMED", order.Status)
	}

	balance, err := getBalance.Execute(ctx, userID)
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if balance != amount {
		t.Fatalf("wallet balance = %d, want %d", balance, amount)
	}
}
