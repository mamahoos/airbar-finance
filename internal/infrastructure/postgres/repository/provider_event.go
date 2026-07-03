package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// CountByPaymentOrderID returns provider events linked to a payment order.
func (r *ProviderEventRepository) CountByPaymentOrderID(ctx context.Context, paymentOrderID string) (int64, error) {
	var count int64
	err := r.querier(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM finance.provider_events
		WHERE payment_order_id = $1
	`, paymentOrderID).Scan(&count)
	return count, err
}

// List returns provider events ordered newest-first.
func (r *ProviderEventRepository) List(ctx context.Context, filter domainprovider.ListFilter) ([]domainprovider.Event, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	where := `WHERE ($1 = '' OR provider = $1)
		AND ($2 = '' OR event_type = $2)
		AND ($3 = '' OR payment_order_id::text = $3)`
	args := []any{filter.Provider, string(filter.EventType), filter.PaymentOrderID}

	var total int64
	if err := r.querier(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM finance.provider_events `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.querier(ctx).Query(ctx, `
		SELECT id, provider, event_type, COALESCE(payment_order_id::text, ''), payload_hash,
			idempotency_key, processed, created_at
		FROM finance.provider_events
		`+where+`
		ORDER BY created_at DESC
		OFFSET $4 LIMIT $5
	`, append(args, offset, limit)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	events := make([]domainprovider.Event, 0)
	for rows.Next() {
		event, err := scanProviderEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

type providerEventScanner interface {
	Scan(dest ...any) error
}

func scanProviderEvent(row providerEventScanner) (domainprovider.Event, error) {
	var event domainprovider.Event
	var eventType string
	err := row.Scan(
		&event.ID,
		&event.Provider,
		&eventType,
		&event.PaymentOrderID,
		&event.PayloadHash,
		&event.IdempotencyKey,
		&event.Processed,
		&event.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domainprovider.Event{}, err
		}
		return domainprovider.Event{}, err
	}
	event.EventType = domainprovider.EventType(eventType)
	return event, nil
}

func (r *ProviderEventRepository) querier(ctx context.Context) pgxQuerier {
	if tx, ok := pg.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}
