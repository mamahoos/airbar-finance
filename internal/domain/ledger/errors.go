package ledger

import "errors"

var (
	// ErrUnbalancedJournal is returned when total debits != total credits.
	ErrUnbalancedJournal = errors.New("ledger: debits must equal credits")
	// ErrEmptyJournal is returned when a journal has no entry lines.
	ErrEmptyJournal = errors.New("ledger: journal must have at least one entry")
	// ErrInvalidEntry is returned when an entry line has invalid debit/credit sides.
	ErrInvalidEntry = errors.New("ledger: each entry must have exactly one of debit or credit")
	// ErrDuplicateJournal is returned when ref_type+ref_id already exists.
	ErrDuplicateJournal = errors.New("ledger: journal already exists for ref")
)
