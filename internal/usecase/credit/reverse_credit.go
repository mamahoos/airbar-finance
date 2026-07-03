package credit

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	domaincredit "github.com/mamahoos/airbar-finance/internal/domain/credit"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
)

// ReverseCreditInput is the application input for reversing a grant.
type ReverseCreditInput struct {
	GrantID       string
	ReverseReason string
	ReversedBy    string
}

// ReverseCreditGrant reverses an active promo credit grant if not consumed.
type ReverseCreditGrant struct {
	pool        *pgxpool.Pool
	credits     domaincredit.Repository
	postJournal *ledgeruc.PostJournal
	audit       *audituc.Emitter
}

// NewReverseCreditGrant creates the ReverseCreditGrant use case.
func NewReverseCreditGrant(
	pool *pgxpool.Pool,
	credits domaincredit.Repository,
	postJournal *ledgeruc.PostJournal,
	audit *audituc.Emitter,
) *ReverseCreditGrant {
	return &ReverseCreditGrant{
		pool:        pool,
		credits:     credits,
		postJournal: postJournal,
		audit:       audit,
	}
}

// Execute reverses the full grant amount via a balancing journal entry.
func (uc *ReverseCreditGrant) Execute(ctx context.Context, input ReverseCreditInput) (*domaincredit.Grant, error) {
	if input.GrantID == "" || input.ReverseReason == "" || input.ReversedBy == "" {
		return nil, domaincredit.ErrInvalidInput
	}

	grant, err := uc.credits.GetGrantByID(ctx, input.GrantID)
	if err != nil {
		return nil, err
	}
	if grant.Status == domaincredit.StatusReversed {
		return nil, domaincredit.ErrAlreadyReversed
	}
	if grant.Status != domaincredit.StatusActive {
		return nil, domaincredit.ErrInvalidInput
	}

	reversedAt := time.Now().UTC()
	var result *domaincredit.Grant
	err = pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		if err := uc.credits.MarkReversed(txCtx, grant.ID, reversedAt, input.ReverseReason, input.ReversedBy); err != nil {
			return err
		}

		_, err := uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypePromoCreditReverse,
			RefID:       fmt.Sprintf("%s:reverse", grant.ID),
			Description: fmt.Sprintf("Promo credit reverse: %s", input.ReverseReason),
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.UserPromoCreditAccount(grant.UserID), Debit: grant.AmountRials, Credit: 0},
				{AccountCode: domainledger.AccountAirbarPromoExpense, Debit: 0, Credit: grant.AmountRials},
			},
		})
		if err != nil {
			return err
		}

		_ = uc.audit.EmitCreditReversed(txCtx, grant.ID, grant.UserID, grant.AmountRials, input.ReverseReason)
		grant.Status = domaincredit.StatusReversed
		grant.ReversedAt = &reversedAt
		grant.ReverseReason = input.ReverseReason
		grant.ReversedBy = input.ReversedBy
		result = grant
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
