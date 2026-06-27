package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	domainprovider "github.com/mamahoos/airbar-finance/internal/domain/provider"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
)

// ProviderEventRepository implements domainprovider.Repository using PostgreSQL.
type ProviderEventRepository struct {
	pool *pgxpool.Pool
}

// NewProviderEventRepository creates a Postgres-backed provider event repository.
func NewProviderEventRepository(pool *pgxpool.Pool) *ProviderEventRepository {
	return &ProviderEventRepository{pool: pool}
}

// Create inserts a provider audit event.
func (r *ProviderEventRepository) Create(ctx context.Context, event *domainprovider.Event) error {
	id := uuid.NewString()
	var createdAt time.Time

	var paymentOrderID any
	if event.PaymentOrderID != "" {
		paymentOrderID = event.PaymentOrderID
	}

	err := r.querier(ctx).QueryRow(ctx, `
		INSERT INTO finance.provider_events (
			id, provider, event_type, payment_order_id, payload, payload_hash, idempotency_key, processed
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`, id, event.Provider, string(event.EventType), paymentOrderID, event.Payload, event.PayloadHash,
		event.IdempotencyKey, event.Processed).Scan(&createdAt)
	if err != nil {
		return err
	}

	event.ID = id
	event.CreatedAt = createdAt
	return nil
}

func (r *ProviderEventRepository) querier(ctx context.Context) pgxQuerier {
	if tx, ok := pg.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}
