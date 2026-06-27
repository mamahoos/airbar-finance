//go:build integration

package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	domainidempotency "github.com/mamahoos/airbar-finance/internal/domain/idempotency"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
)

func TestIdempotencyRepositoryIntegration(t *testing.T) {
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

	repo := NewIdempotencyRepository(pool)
	key := "integration-key-" + uuid.NewString()

	acquired, existing, err := repo.TryBeginProcessing(ctx, &domainidempotency.Record{
		Key:          key,
		Scope:        "escrow.create",
		ResourceType: "shipment",
		ResourceID:   "sh-test",
	})
	if err != nil {
		t.Fatalf("TryBeginProcessing() error = %v", err)
	}
	if !acquired || existing == nil {
		t.Fatal("expected first TryBeginProcessing to acquire lock")
	}

	acquired, existing, err = repo.TryBeginProcessing(ctx, &domainidempotency.Record{Key: key})
	if err != nil {
		t.Fatalf("TryBeginProcessing() duplicate error = %v", err)
	}
	if acquired {
		t.Fatal("expected duplicate TryBeginProcessing to return existing row")
	}
	if existing.Status != domainidempotency.StatusProcessing {
		t.Fatalf("status = %q, want PROCESSING", existing.Status)
	}

	snapshot := map[string]any{"status": "CREATED"}
	if err := repo.Complete(ctx, key, snapshot); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	stored, err := repo.GetByKey(ctx, key)
	if err != nil {
		t.Fatalf("GetByKey() error = %v", err)
	}
	if stored.Status != domainidempotency.StatusCompleted {
		t.Fatalf("status = %q, want COMPLETED", stored.Status)
	}
	if stored.ResponseSnapshot["status"] != "CREATED" {
		t.Fatalf("snapshot = %#v", stored.ResponseSnapshot)
	}

	acquired, existing, err = repo.TryBeginProcessing(ctx, &domainidempotency.Record{Key: key})
	if err != nil {
		t.Fatalf("TryBeginProcessing() completed error = %v", err)
	}
	if acquired || existing.Status != domainidempotency.StatusCompleted {
		t.Fatalf("expected completed replay, acquired=%v status=%q", acquired, existing.Status)
	}
}
