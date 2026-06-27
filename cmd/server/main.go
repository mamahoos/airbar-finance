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
	escrowuc "github.com/mamahoos/airbar-finance/internal/usecase/escrow"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
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

	ensureWallet := walletuc.NewEnsureWalletAccount(walletRepo)
	postJournal := ledgeruc.NewPostJournal(ledgerRepo, ensureWallet)
	getBalance := walletuc.NewGetBalance(ledgerRepo)

	createEscrow := escrowuc.NewCreateEscrow(escrowRepo)
	getEscrow := escrowuc.NewGetEscrow(escrowRepo)
	fundEscrow := escrowuc.NewFundEscrow(dbPool, escrowRepo, postJournal)
	payFromWallet := escrowuc.NewPayFromWallet(dbPool, escrowRepo, postJournal, getBalance)
	markDelivered := escrowuc.NewMarkDelivered(escrowRepo)
	freezeEscrow := escrowuc.NewFreezeEscrow(escrowRepo)
	releaseEscrow := escrowuc.NewReleaseEscrow(dbPool, escrowRepo, postJournal, ledgerRepo, cfg.PlatformFeePercent)
	refundEscrow := escrowuc.NewRefundEscrow(dbPool, escrowRepo, postJournal, ledgerRepo)
	partialRefundEscrow := escrowuc.NewPartialRefundEscrow(dbPool, escrowRepo, postJournal, ledgerRepo)

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

	grpcServer, err := deliverygrpc.NewServer(cfg.GRPCPort, checker, escrowHandler)
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: newHTTPHandler(checker),
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

func newHTTPHandler(checker *health.Checker) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/health/ready", deliveryhttp.NewHealthHandler(checker))
	return mux
}
