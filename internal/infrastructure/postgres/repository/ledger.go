package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

// LedgerRepository implements domainledger.Repository using PostgreSQL.
type LedgerRepository struct {
	pool *pgxpool.Pool
}

// NewLedgerRepository creates a Postgres-backed ledger repository.
func NewLedgerRepository(pool *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{pool: pool}
}

// CreateJournal inserts a journal and its entries in a single transaction.
func (r *LedgerRepository) CreateJournal(ctx context.Context, journal *domainledger.Journal) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	journalID := uuid.NewString()

	var createdAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO finance.ledger_journals (id, ref_type, ref_id, description)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`, journalID, string(journal.RefType), journal.RefID, journal.Description).Scan(&createdAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domainledger.ErrDuplicateJournal
		}
		return err
	}

	for i := range journal.Entries {
		entryID := uuid.NewString()
		line := journal.Entries[i]
		_, err = tx.Exec(ctx, `
			INSERT INTO finance.ledger_entries (id, journal_id, account_code, debit, credit)
			VALUES ($1, $2, $3, $4, $5)
		`, entryID, journalID, line.AccountCode.String(), line.Debit, line.Credit)
		if err != nil {
			return err
		}

		journal.Entries[i].ID = entryID
		journal.Entries[i].JournalID = journalID
	}

	if err := tx.Commit(ctx); err != nil {
		if isUniqueViolation(err) {
			return domainledger.ErrDuplicateJournal
		}
		return err
	}

	journal.ID = journalID
	journal.CreatedAt = createdAt
	return nil
}

// SumByAccount returns total debits and credits posted to an account.
func (r *LedgerRepository) SumByAccount(ctx context.Context, accountCode domainledger.AccountCode) (int64, int64, error) {
	var debit int64
	var credit int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(debit), 0), COALESCE(SUM(credit), 0)
		FROM finance.ledger_entries
		WHERE account_code = $1
	`, accountCode.String()).Scan(&debit, &credit)
	if err != nil {
		return 0, 0, err
	}
	return debit, credit, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
