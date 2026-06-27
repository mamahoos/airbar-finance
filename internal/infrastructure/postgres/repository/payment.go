package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
)

// PaymentRepository implements domainpayment.Repository using PostgreSQL.
type PaymentRepository struct {
	pool *pgxpool.Pool
}

// NewPaymentRepository creates a Postgres-backed payment repository.
func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

// Create inserts a payment order.
func (r *PaymentRepository) Create(ctx context.Context, order *domainpayment.Order) error {
	id := uuid.NewString()
	var createdAt time.Time
	var updatedAt time.Time

	var agreedPrice any
	if order.AgreedPrice > 0 {
		agreedPrice = order.AgreedPrice
	}

	err := r.querier(ctx).QueryRow(ctx, `
		INSERT INTO finance.payment_orders (
			id, shipment_id, payer_user_id, purpose, amount, status,
			authority, redirect_url, success_url, failure_url, description, agreed_price
		) VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), $9, $10, $11, $12)
		RETURNING created_at, updated_at
	`, id, order.ShipmentID, order.PayerUserID, string(order.Purpose), order.Amount, string(order.Status),
		order.Authority, order.RedirectURL, order.SuccessURL, order.FailureURL, order.Description, agreedPrice).
		Scan(&createdAt, &updatedAt)
	if err != nil {
		return err
	}

	order.ID = id
	order.CreatedAt = createdAt
	order.UpdatedAt = updatedAt
	return nil
}

// GetByID loads a payment order by id.
func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*domainpayment.Order, error) {
	row := r.querier(ctx).QueryRow(ctx, paymentOrderSelectSQL+` WHERE id = $1`, id)
	return scanPaymentOrder(row)
}

// GetByAuthority loads a payment order by Zibal trackId.
func (r *PaymentRepository) GetByAuthority(ctx context.Context, authority string) (*domainpayment.Order, error) {
	row := r.querier(ctx).QueryRow(ctx, paymentOrderSelectSQL+` WHERE authority = $1`, authority)
	return scanPaymentOrder(row)
}

// Update persists payment order changes.
func (r *PaymentRepository) Update(ctx context.Context, order *domainpayment.Order) error {
	tag, err := r.querier(ctx).Exec(ctx, `
		UPDATE finance.payment_orders
		SET status = $2,
		    authority = NULLIF($3, ''),
		    redirect_url = NULLIF($4, ''),
		    verified_at = $5,
		    updated_at = now()
		WHERE id = $1
	`, order.ID, string(order.Status), order.Authority, order.RedirectURL, order.VerifiedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domainpayment.ErrNotFound
	}
	return nil
}

const paymentOrderSelectSQL = `
	SELECT id,
	       COALESCE(shipment_id, ''),
	       payer_user_id,
	       purpose,
	       amount,
	       status,
	       COALESCE(authority, ''),
	       COALESCE(redirect_url, ''),
	       success_url,
	       failure_url,
	       description,
	       COALESCE(agreed_price, 0),
	       verified_at,
	       created_at,
	       updated_at
	FROM finance.payment_orders
`

func scanPaymentOrder(row rowScanner) (*domainpayment.Order, error) {
	var order domainpayment.Order
	var purpose string
	var status string
	var verifiedAt *time.Time

	err := row.Scan(
		&order.ID,
		&order.ShipmentID,
		&order.PayerUserID,
		&purpose,
		&order.Amount,
		&status,
		&order.Authority,
		&order.RedirectURL,
		&order.SuccessURL,
		&order.FailureURL,
		&order.Description,
		&order.AgreedPrice,
		&verifiedAt,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainpayment.ErrNotFound
		}
		return nil, err
	}

	order.Purpose = domainpayment.Purpose(purpose)
	order.Status = domainpayment.Status(status)
	order.VerifiedAt = verifiedAt
	return &order, nil
}

func (r *PaymentRepository) querier(ctx context.Context) pgxQuerier {
	if tx, ok := pg.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func isPaymentUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
