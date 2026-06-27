package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
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
	if tx, ok := pg.TxFromContext(ctx); ok {
		return r.createJournal(ctx, tx, journal)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.createJournal(ctx, tx, journal); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *LedgerRepository) createJournal(ctx context.Context, tx pgx.Tx, journal *domainledger.Journal) error {
	journalID := uuid.NewString()

	var createdAt time.Time
	err := tx.QueryRow(ctx, `
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

// ListByAccount returns ledger lines for an account joined with journal metadata.
func (r *LedgerRepository) ListByAccount(ctx context.Context, accountCode domainledger.AccountCode) ([]domainledger.AccountEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT j.id, j.ref_type, j.ref_id, j.description, e.debit, e.credit, j.created_at
		FROM finance.ledger_entries e
		INNER JOIN finance.ledger_journals j ON j.id = e.journal_id
		WHERE e.account_code = $1
		ORDER BY j.created_at DESC, e.created_at DESC
	`, accountCode.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domainledger.AccountEntry
	for rows.Next() {
		var entry domainledger.AccountEntry
		var refType string
		if err := rows.Scan(
			&entry.JournalID,
			&refType,
			&entry.RefID,
			&entry.Description,
			&entry.Debit,
			&entry.Credit,
			&entry.CreatedAt,
		); err != nil {
			return nil, err
		}
		entry.RefType = domainledger.RefType(refType)
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
