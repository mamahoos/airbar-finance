//go:build integration

package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	domainaudit "github.com/mamahoos/airbar-finance/internal/domain/audit"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	escrowuc "github.com/mamahoos/airbar-finance/internal/usecase/escrow"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

func TestFinanceAuditEscrowCreateIntegration(t *testing.T) {
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

	escrowRepo := NewEscrowRepository(pool)
	financeEventRepo := NewFinanceEventRepository(pool)
	auditEmitter := audituc.NewEmitter(financeEventRepo)
	createEscrow := escrowuc.NewCreateEscrow(escrowRepo, auditEmitter)

	suffix := uuid.NewString()[:8]
	escrow, err := createEscrow.Execute(ctx, escrowuc.CreateEscrowInput{
		ShipmentID:    "audit-sh-" + suffix,
		CarrierUserID: "carrier-" + suffix,
		PayerUserID:   "payer-" + suffix,
		Amount:        12000,
	})
	if err != nil {
		t.Fatalf("CreateEscrow() error = %v", err)
	}

	count, err := financeEventRepo.CountByAggregate(ctx, domainaudit.AggregateEscrow, escrow.ID)
	if err != nil {
		t.Fatalf("CountByAggregate() error = %v", err)
	}
	if count < 1 {
		t.Fatalf("finance events = %d, want at least 1", count)
	}
}
