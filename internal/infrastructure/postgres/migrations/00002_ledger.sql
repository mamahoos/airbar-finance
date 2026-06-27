-- +goose Up
-- F1: double-entry ledger — journals (header) and entries (lines).
CREATE TABLE finance.ledger_journals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ref_type TEXT NOT NULL,
    ref_id TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ledger_journals_ref_unique UNIQUE (ref_type, ref_id)
);

CREATE TABLE finance.ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id UUID NOT NULL REFERENCES finance.ledger_journals (id) ON DELETE RESTRICT,
    account_code TEXT NOT NULL,
    debit BIGINT NOT NULL DEFAULT 0 CHECK (debit >= 0),
    credit BIGINT NOT NULL DEFAULT 0 CHECK (credit >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ledger_entries_side_check CHECK (
        (debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0)
    )
);

CREATE INDEX idx_ledger_entries_journal_id ON finance.ledger_entries (journal_id);
CREATE INDEX idx_ledger_entries_account_code ON finance.ledger_entries (account_code);

-- +goose Down
DROP TABLE IF EXISTS finance.ledger_entries;
DROP TABLE IF EXISTS finance.ledger_journals;
