package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
)

// WithdrawalRepository implements domainwithdrawal.Repository using PostgreSQL.
type WithdrawalRepository struct {
	pool *pgxpool.Pool
}

// NewWithdrawalRepository creates a Postgres-backed withdrawal repository.
func NewWithdrawalRepository(pool *pgxpool.Pool) *WithdrawalRepository {
	return &WithdrawalRepository{pool: pool}
}

// Create inserts a withdrawal row.
func (r *WithdrawalRepository) Create(ctx context.Context, withdrawal *domainwithdrawal.Withdrawal) error {
	id := uuid.NewString()
	var createdAt time.Time
	var updatedAt time.Time

	err := r.querier(ctx).QueryRow(ctx, `
		INSERT INTO finance.withdrawals (id, user_id, amount, status, destination_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`, id, withdrawal.UserID, withdrawal.Amount, string(withdrawal.Status), withdrawal.DestinationHash).
		Scan(&createdAt, &updatedAt)
	if err != nil {
		return err
	}

	withdrawal.ID = id
	withdrawal.CreatedAt = createdAt
	withdrawal.UpdatedAt = updatedAt
	return nil
}

// GetByID loads a withdrawal by id.
func (r *WithdrawalRepository) GetByID(ctx context.Context, id string) (*domainwithdrawal.Withdrawal, error) {
	row := r.querier(ctx).QueryRow(ctx, withdrawalSelectSQL+` WHERE id = $1`, id)
	return scanWithdrawal(row)
}

// List returns withdrawals filtered by user and optional status.
func (r *WithdrawalRepository) List(ctx context.Context, userID string, status domainwithdrawal.Status) ([]domainwithdrawal.Withdrawal, error) {
	query := withdrawalSelectSQL + ` WHERE user_id = $1`
	args := []any{userID}
	if status != "" {
		query += ` AND status = $2`
		args = append(args, string(status))
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.querier(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domainwithdrawal.Withdrawal
	for rows.Next() {
		item, err := scanWithdrawal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// Update persists withdrawal changes.
func (r *WithdrawalRepository) Update(ctx context.Context, withdrawal *domainwithdrawal.Withdrawal) error {
	tag, err := r.querier(ctx).Exec(ctx, `
		UPDATE finance.withdrawals
		SET status = $2,
		    provider_ref = NULLIF($3, ''),
		    reject_reason = NULLIF($4, ''),
		    processed_at = $5,
		    updated_at = now()
		WHERE id = $1
	`, withdrawal.ID, string(withdrawal.Status), withdrawal.ProviderRef, withdrawal.RejectReason, withdrawal.ProcessedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domainwithdrawal.ErrNotFound
	}
	return nil
}

const withdrawalSelectSQL = `
	SELECT id, user_id, amount, status,
	       destination_hash,
	       COALESCE(provider_ref, ''),
	       COALESCE(reject_reason, ''),
	       processed_at, created_at, updated_at
	FROM finance.withdrawals
`

func scanWithdrawal(row rowScanner) (*domainwithdrawal.Withdrawal, error) {
	var item domainwithdrawal.Withdrawal
	var status string
	var processedAt *time.Time

	err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Amount,
		&status,
		&item.DestinationHash,
		&item.ProviderRef,
		&item.RejectReason,
		&processedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainwithdrawal.ErrNotFound
		}
		return nil, err
	}

	item.Status = domainwithdrawal.Status(status)
	item.ProcessedAt = processedAt
	return &item, nil
}

func (r *WithdrawalRepository) querier(ctx context.Context) pgxQuerier {
	if tx, ok := pg.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}
