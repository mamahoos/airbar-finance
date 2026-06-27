package ledger

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

// PostJournalInput is the application input for posting a balanced journal.
type PostJournalInput struct {
	RefType     domainledger.RefType
	RefID       string
	Description string
	Lines       []domainledger.EntryLine
}

// PostJournal posts a double-entry journal atomically via the ledger repository.
type PostJournal struct {
	repo domainledger.Repository
}

// NewPostJournal creates the PostJournal use case.
func NewPostJournal(repo domainledger.Repository) *PostJournal {
	return &PostJournal{repo: repo}
}

// Execute validates invariants and persists the journal.
func (uc *PostJournal) Execute(ctx context.Context, input PostJournalInput) (*domainledger.Journal, error) {
	journal, err := domainledger.NewJournal(input.RefType, input.RefID, input.Description, input.Lines)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.CreateJournal(ctx, journal); err != nil {
		return nil, err
	}

	return journal, nil
}
