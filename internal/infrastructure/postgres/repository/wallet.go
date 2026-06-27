package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainwallet "github.com/mamahoos/airbar-finance/internal/domain/wallet"
)

// WalletRepository implements domainwallet.Repository using PostgreSQL.
type WalletRepository struct {
	pool *pgxpool.Pool
}

// NewWalletRepository creates a Postgres-backed wallet repository.
func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

// EnsureAccount lazily registers a wallet account for the user (no balance column).
func (r *WalletRepository) EnsureAccount(ctx context.Context, userID string) (*domainwallet.Account, error) {
	accountCode := domainledger.UserWalletAccount(userID).String()

	var account domainwallet.Account
	err := r.pool.QueryRow(ctx, `
		INSERT INTO finance.wallet_accounts (id, user_id, currency, account_code)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, currency) DO UPDATE
		SET user_id = finance.wallet_accounts.user_id
		RETURNING id, user_id, currency, account_code, created_at
	`, uuid.NewString(), userID, domainwallet.CurrencyIRT, accountCode).Scan(
		&account.ID,
		&account.UserID,
		&account.Currency,
		&account.AccountCode,
		&account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetByUserID returns the registered wallet account if it exists.
func (r *WalletRepository) GetByUserID(ctx context.Context, userID string) (*domainwallet.Account, error) {
	var account domainwallet.Account
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, currency, account_code, created_at
		FROM finance.wallet_accounts
		WHERE user_id = $1 AND currency = $2
	`, userID, domainwallet.CurrencyIRT).Scan(
		&account.ID,
		&account.UserID,
		&account.Currency,
		&account.AccountCode,
		&account.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &account, nil
}
