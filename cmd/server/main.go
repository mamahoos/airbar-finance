package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliverygrpc "github.com/mamahoos/airbar-finance/internal/delivery/grpc"
	"github.com/mamahoos/airbar-finance/internal/delivery/grpc/handlers"
	deliveryhttp "github.com/mamahoos/airbar-finance/internal/delivery/http"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/config"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/health"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres/repository"
	redisinfra "github.com/mamahoos/airbar-finance/internal/infrastructure/redis"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/zibal"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
	credituc "github.com/mamahoos/airbar-finance/internal/usecase/credit"
	escrowuc "github.com/mamahoos/airbar-finance/internal/usecase/escrow"
	idempotencyuc "github.com/mamahoos/airbar-finance/internal/usecase/idempotency"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	paymentuc "github.com/mamahoos/airbar-finance/internal/usecase/payment"
	provideruc "github.com/mamahoos/airbar-finance/internal/usecase/provider"
	reconuc "github.com/mamahoos/airbar-finance/internal/usecase/reconciliation"
	treasuryuc "github.com/mamahoos/airbar-finance/internal/usecase/treasury"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
	withdrawaluc "github.com/mamahoos/airbar-finance/internal/usecase/withdrawal"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	redisClient, err := redisinfra.NewClient(cfg.RedisURL)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	checker := health.NewChecker(dbPool, redisClient)

	ledgerRepo := repository.NewLedgerRepository(dbPool)
	walletRepo := repository.NewWalletRepository(dbPool)
	escrowRepo := repository.NewEscrowRepository(dbPool)
	paymentRepo := repository.NewPaymentRepository(dbPool)
	providerEventRepo := repository.NewProviderEventRepository(dbPool)

	withdrawalRepo := repository.NewWithdrawalRepository(dbPool)
	creditRepo := repository.NewCreditRepository(dbPool)
	idempotencyRepo := repository.NewIdempotencyRepository(dbPool)
	idempotencyCache := redisinfra.NewIdempotencyCache(redisClient)
	idempotencyGuard := idempotencyuc.NewGuard(idempotencyRepo, idempotencyCache)
	financeEventRepo := repository.NewFinanceEventRepository(dbPool)
	auditEmitter := audituc.NewEmitter(financeEventRepo)

	zibalClient := zibal.NewClient(cfg)

	ensureWallet := walletuc.NewEnsureWalletAccount(walletRepo)
	ensureCredit := credituc.NewEnsureCreditAccount(creditRepo)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, ensureWallet)
	getBalance := walletuc.NewGetBalance(ledgerRepo)
	getWallet := walletuc.NewGetWallet(getBalance)
	listWalletTransactions := walletuc.NewListWalletTransactions(ledgerRepo)

	createEscrow := escrowuc.NewCreateEscrow(escrowRepo, auditEmitter)
	getEscrow := escrowuc.NewGetEscrow(escrowRepo)
	fundEscrow := escrowuc.NewFundEscrow(dbPool, escrowRepo, postJournal, auditEmitter)
	payFromWallet := escrowuc.NewPayFromWallet(dbPool, escrowRepo, postJournal, getBalance, auditEmitter)
	markDelivered := escrowuc.NewMarkDelivered(escrowRepo, auditEmitter)
	freezeEscrow := escrowuc.NewFreezeEscrow(escrowRepo, auditEmitter)
	releaseEscrow := escrowuc.NewReleaseEscrow(dbPool, escrowRepo, postJournal, ledgerRepo, cfg.PlatformFeePercent, auditEmitter)
	refundEscrow := escrowuc.NewRefundEscrow(dbPool, escrowRepo, postJournal, ledgerRepo, auditEmitter)
	partialRefundEscrow := escrowuc.NewPartialRefundEscrow(dbPool, escrowRepo, postJournal, ledgerRepo, auditEmitter)

	verifyOrder := paymentuc.NewVerifyOrder(dbPool, paymentRepo, providerEventRepo, zibalClient, fundEscrow, postJournal, auditEmitter)
	createPaymentOrder := paymentuc.NewCreatePaymentOrder(paymentRepo, escrowRepo, providerEventRepo, zibalClient, cfg.FinancePublicBaseURL, auditEmitter)
	getPaymentOrder := paymentuc.NewGetPaymentOrder(paymentRepo)
	verifyPaymentOrder := paymentuc.NewVerifyPaymentOrder(verifyOrder)
	createWalletTopupOrder := paymentuc.NewCreateWalletTopupOrder(paymentRepo, providerEventRepo, zibalClient, cfg.FinancePublicBaseURL, auditEmitter)
	verifyWalletTopupOrder := paymentuc.NewVerifyWalletTopupOrder(verifyOrder)
	failPaymentOrder := paymentuc.NewFailPaymentOrder(paymentRepo, providerEventRepo, auditEmitter)
	handleCallback := paymentuc.NewHandleCallback(verifyOrder, failPaymentOrder, providerEventRepo)

	escrowHandler := handlers.NewEscrowHandler(
		createEscrow,
		getEscrow,
		fundEscrow,
		payFromWallet,
		markDelivered,
		freezeEscrow,
		releaseEscrow,
		refundEscrow,
		partialRefundEscrow,
	)

	paymentHandler := handlers.NewPaymentHandler(
		createPaymentOrder,
		getPaymentOrder,
		verifyPaymentOrder,
		createWalletTopupOrder,
		verifyWalletTopupOrder,
	)

	walletHandler := handlers.NewWalletHandler(getWallet, listWalletTransactions)

	createWithdrawal := withdrawaluc.NewCreateWithdrawal(dbPool, withdrawalRepo, postJournal, getBalance, auditEmitter)
	listWithdrawals := withdrawaluc.NewListWithdrawals(withdrawalRepo)
	approveWithdrawal := withdrawaluc.NewApproveWithdrawal(withdrawalRepo, auditEmitter)
	markWithdrawalSent := withdrawaluc.NewMarkWithdrawalSent(withdrawalRepo, auditEmitter)
	settleWithdrawal := withdrawaluc.NewSettleWithdrawal(withdrawalRepo, auditEmitter)
	failWithdrawal := withdrawaluc.NewFailWithdrawal(dbPool, withdrawalRepo, postJournal, auditEmitter)
	processWithdrawal := withdrawaluc.NewProcessWithdrawal(withdrawalRepo, auditEmitter)
	rejectWithdrawal := withdrawaluc.NewRejectWithdrawal(dbPool, withdrawalRepo, postJournal, auditEmitter)

	getTreasurySummary := treasuryuc.NewGetTreasurySummary(ledgerRepo)
	reconciliationRepo := repository.NewReconciliationRepository(dbPool)
	runReconciliation := reconuc.NewRunReconciliation(ledgerRepo, reconciliationRepo)
	listReconciliationRuns := reconuc.NewListReconciliationRuns(reconciliationRepo)
	getReconciliationRun := reconuc.NewGetReconciliationRun(reconciliationRepo)
	listProviderEvents := provideruc.NewListProviderEvents(providerEventRepo)

	withdrawalHandler := handlers.NewWithdrawalHandler(
		createWithdrawal,
		listWithdrawals,
		approveWithdrawal,
		markWithdrawalSent,
		settleWithdrawal,
		failWithdrawal,
		processWithdrawal,
		rejectWithdrawal,
	)

	treasuryHandler := handlers.NewTreasuryHandler(getTreasurySummary)
	reconciliationHandler := handlers.NewReconciliationHandler(
		runReconciliation,
		listReconciliationRuns,
		getReconciliationRun,
	)
	providerEventHandler := handlers.NewProviderEventHandler(listProviderEvents)

	getCreditBalance := credituc.NewGetBalance(ledgerRepo)
	listCreditGrants := credituc.NewListGrants(creditRepo, getCreditBalance)
	grantCredit := credituc.NewGrantCredit(dbPool, creditRepo, ensureCredit, postJournal, auditEmitter)
	reverseCreditGrant := credituc.NewReverseCreditGrant(dbPool, creditRepo, postJournal, auditEmitter)
	creditHandler := handlers.NewCreditHandler(grantCredit, reverseCreditGrant, getCreditBalance, listCreditGrants)

	grpcServer, err := deliverygrpc.NewServer(
		cfg.GRPCPort,
		checker,
		idempotencyGuard,
		escrowHandler,
		paymentHandler,
		walletHandler,
		withdrawalHandler,
		treasuryHandler,
		reconciliationHandler,
		providerEventHandler,
		creditHandler,
	)
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: newHTTPHandler(checker, handleCallback),
	}

	errCh := make(chan error, 2)
	go func() {
		logger.Info("gRPC server listening", slog.String("addr", grpcServer.Addr().String()))
		errCh <- grpcServer.Serve()
	}()
	go func() {
		logger.Info("HTTP server listening", slog.String("addr", httpServer.Addr))
		errCh <- httpServer.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("shutdown complete")
	return nil
}

func newHTTPHandler(checker *health.Checker, handleCallback *paymentuc.HandleCallback) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/health/ready", deliveryhttp.NewHealthHandler(checker))
	mux.Handle("/api/v1/zibal/callback", deliveryhttp.NewZibalCallbackHandler(handleCallback))
	return mux
}
