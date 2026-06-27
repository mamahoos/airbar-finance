package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
)

// EscrowRepository implements domainescrow.Repository using PostgreSQL.
type EscrowRepository struct {
	pool *pgxpool.Pool
}

// NewEscrowRepository creates a Postgres-backed escrow repository.
func NewEscrowRepository(pool *pgxpool.Pool) *EscrowRepository {
	return &EscrowRepository{pool: pool}
}

// Create inserts a new escrow row.
func (r *EscrowRepository) Create(ctx context.Context, escrow *domainescrow.Escrow) error {
	return r.create(ctx, r.querier(ctx), escrow)
}

// GetByShipmentID loads an escrow by shipment id.
func (r *EscrowRepository) GetByShipmentID(ctx context.Context, shipmentID string) (*domainescrow.Escrow, error) {
	row := r.querier(ctx).QueryRow(ctx, `
		SELECT id, shipment_id, carrier_user_id, payer_user_id, amount, status,
		       COALESCE(payment_order_id, ''), COALESCE(funding_source, ''),
		       funded_at, released_at, refunded_at, created_at, updated_at
		FROM finance.escrows
		WHERE shipment_id = $1
	`, shipmentID)

	escrow, err := scanEscrow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainescrow.ErrNotFound
		}
		return nil, err
	}
	return escrow, nil
}

// Update persists escrow state changes.
func (r *EscrowRepository) Update(ctx context.Context, escrow *domainescrow.Escrow) error {
	tag, err := r.querier(ctx).Exec(ctx, `
		UPDATE finance.escrows
		SET status = $2,
		    payment_order_id = NULLIF($3, ''),
		    funding_source = NULLIF($4, ''),
		    funded_at = $5,
		    released_at = $6,
		    refunded_at = $7,
		    updated_at = now()
		WHERE id = $1
	`, escrow.ID, string(escrow.Status), escrow.PaymentOrderID, string(escrow.FundingSource),
		escrow.FundedAt, escrow.ReleasedAt, escrow.RefundedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domainescrow.ErrNotFound
	}
	return nil
}

func (r *EscrowRepository) create(ctx context.Context, q pgxQuerier, escrow *domainescrow.Escrow) error {
	id := uuid.NewString()
	var createdAt time.Time
	var updatedAt time.Time
	err := q.QueryRow(ctx, `
		INSERT INTO finance.escrows (
			id, shipment_id, carrier_user_id, payer_user_id, amount, status
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, id, escrow.ShipmentID, escrow.CarrierUserID, escrow.PayerUserID, escrow.Amount, string(escrow.Status)).
		Scan(&createdAt, &updatedAt)
	if err != nil {
		if isEscrowUniqueViolation(err) {
			return domainescrow.ErrDuplicateShipment
		}
		return err
	}

	escrow.ID = id
	escrow.CreatedAt = createdAt
	escrow.UpdatedAt = updatedAt
	return nil
}

type pgxQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (r *EscrowRepository) querier(ctx context.Context) pgxQuerier {
	if tx, ok := pg.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEscrow(row rowScanner) (*domainescrow.Escrow, error) {
	var escrow domainescrow.Escrow
	var status string
	var paymentOrderID string
	var fundingSource string
	var fundedAt *time.Time
	var releasedAt *time.Time
	var refundedAt *time.Time

	err := row.Scan(
		&escrow.ID,
		&escrow.ShipmentID,
		&escrow.CarrierUserID,
		&escrow.PayerUserID,
		&escrow.Amount,
		&status,
		&paymentOrderID,
		&fundingSource,
		&fundedAt,
		&releasedAt,
		&refundedAt,
		&escrow.CreatedAt,
		&escrow.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	escrow.Status = domainescrow.Status(status)
	escrow.PaymentOrderID = paymentOrderID
	escrow.FundingSource = domainescrow.FundingSource(fundingSource)
	escrow.FundedAt = fundedAt
	escrow.ReleasedAt = releasedAt
	escrow.RefundedAt = refundedAt
	return &escrow, nil
}

func isEscrowUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
