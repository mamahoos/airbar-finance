package ledger

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
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
	repo          domainledger.Repository
	walletEnsurer *walletuc.EnsureWalletAccount
}

// NewPostJournal creates the PostJournal use case.
// walletEnsurer is optional; when set, wallet accounts are lazily created on first ledger touch.
func NewPostJournal(repo domainledger.Repository, walletEnsurer *walletuc.EnsureWalletAccount) *PostJournal {
	return &PostJournal{repo: repo, walletEnsurer: walletEnsurer}
}

// Execute validates invariants and persists the journal.
func (uc *PostJournal) Execute(ctx context.Context, input PostJournalInput) (*domainledger.Journal, error) {
	journal, err := domainledger.NewJournal(input.RefType, input.RefID, input.Description, input.Lines)
	if err != nil {
		return nil, err
	}

	if err := walletuc.EnsureForLines(ctx, uc.walletEnsurer, input.Lines); err != nil {
		return nil, err
	}

	if err := uc.repo.CreateJournal(ctx, journal); err != nil {
		return nil, err
	}

	return journal, nil
}
