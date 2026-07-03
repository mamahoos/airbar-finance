package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domaincredit "github.com/mamahoos/airbar-finance/internal/domain/credit"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

// CreditRepository implements domaincredit.Repository using PostgreSQL.
type CreditRepository struct {
	pool *pgxpool.Pool
}

// NewCreditRepository creates a Postgres-backed credit repository.
func NewCreditRepository(pool *pgxpool.Pool) *CreditRepository {
	return &CreditRepository{pool: pool}
}

// EnsureAccount lazily registers a promo credit account for the user.
func (r *CreditRepository) EnsureAccount(ctx context.Context, userID string) (*domaincredit.Account, error) {
	accountCode := domainledger.UserPromoCreditAccount(userID).String()

	var account domaincredit.Account
	err := r.pool.QueryRow(ctx, `
		INSERT INTO finance.credit_accounts (id, user_id, currency, account_code)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, currency) DO UPDATE
		SET user_id = finance.credit_accounts.user_id
		RETURNING id, user_id, currency, account_code, created_at
	`, uuid.NewString(), userID, domaincredit.CurrencyIRT, accountCode).Scan(
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

// CreateGrant persists a promo credit grant row.
func (r *CreditRepository) CreateGrant(ctx context.Context, grant *domaincredit.Grant) error {
	if grant.ID == "" {
		grant.ID = uuid.NewString()
	}
	if grant.Status == "" {
		grant.Status = domaincredit.StatusActive
	}

	var campaignRef *string
	if grant.CampaignRef != "" {
		campaignRef = &grant.CampaignRef
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO finance.credit_grants (
			id, user_id, amount_rials, reason, campaign_ref, expires_at,
			status, granted_by, idempotency_key
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`,
		grant.ID,
		grant.UserID,
		grant.AmountRials,
		grant.Reason,
		campaignRef,
		grant.ExpiresAt,
		string(grant.Status),
		grant.GrantedBy,
		grant.IdempotencyKey,
	).Scan(&grant.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domaincredit.ErrDuplicateGrant
		}
		return err
	}
	return nil
}

// GetGrantByID returns a grant by id.
func (r *CreditRepository) GetGrantByID(ctx context.Context, id string) (*domaincredit.Grant, error) {
	var grant domaincredit.Grant
	var status string
	var campaignRef *string
	var reverseReason *string
	var reversedBy *string
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, amount_rials, reason, campaign_ref, expires_at,
			status, granted_by, idempotency_key, reversed_at, reverse_reason, reversed_by, created_at
		FROM finance.credit_grants
		WHERE id = $1
	`, id).Scan(
		&grant.ID,
		&grant.UserID,
		&grant.AmountRials,
		&grant.Reason,
		&campaignRef,
		&grant.ExpiresAt,
		&status,
		&grant.GrantedBy,
		&grant.IdempotencyKey,
		&grant.ReversedAt,
		&reverseReason,
		&reversedBy,
		&grant.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domaincredit.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if campaignRef != nil {
		grant.CampaignRef = *campaignRef
	}
	if reverseReason != nil {
		grant.ReverseReason = *reverseReason
	}
	if reversedBy != nil {
		grant.ReversedBy = *reversedBy
	}
	grant.Status = domaincredit.GrantStatus(status)
	return &grant, nil
}

// MarkReversed updates a grant to reversed status.
func (r *CreditRepository) MarkReversed(ctx context.Context, id string, reversedAt time.Time, reverseReason, reversedBy string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE finance.credit_grants
		SET status = $2, reversed_at = $3, reverse_reason = $4, reversed_by = $5
		WHERE id = $1 AND status = $6
	`, id, string(domaincredit.StatusReversed), reversedAt, reverseReason, reversedBy, string(domaincredit.StatusActive))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		grant, getErr := r.GetGrantByID(ctx, id)
		if getErr != nil {
			return getErr
		}
		if grant.Status == domaincredit.StatusReversed {
			return domaincredit.ErrAlreadyReversed
		}
		return domaincredit.ErrNotFound
	}
	return nil
}

// ListGrantsByUserID returns grants for a user ordered by newest first.
func (r *CreditRepository) ListGrantsByUserID(ctx context.Context, userID string, limit, offset int) ([]domaincredit.Grant, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount_rials, reason, campaign_ref, expires_at,
			status, granted_by, idempotency_key, reversed_at, reverse_reason, reversed_by, created_at
		FROM finance.credit_grants
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []domaincredit.Grant
	for rows.Next() {
		var grant domaincredit.Grant
		var status string
		var campaignRef *string
		var reverseReason *string
		var reversedBy *string
		if err := rows.Scan(
			&grant.ID,
			&grant.UserID,
			&grant.AmountRials,
			&grant.Reason,
			&campaignRef,
			&grant.ExpiresAt,
			&status,
			&grant.GrantedBy,
			&grant.IdempotencyKey,
			&grant.ReversedAt,
			&reverseReason,
			&reversedBy,
			&grant.CreatedAt,
		); err != nil {
			return nil, err
		}
		if campaignRef != nil {
			grant.CampaignRef = *campaignRef
		}
		if reverseReason != nil {
			grant.ReverseReason = *reverseReason
		}
		if reversedBy != nil {
			grant.ReversedBy = *reversedBy
		}
		grant.Status = domaincredit.GrantStatus(status)
		grants = append(grants, grant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return grants, nil
}
