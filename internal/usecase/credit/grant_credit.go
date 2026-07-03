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

// GrantCreditInput is the application input for UC-25 grant.
type GrantCreditInput struct {
	UserID         string
	Amount         int64
	Reason         string
	CampaignRef    string
	ExpiresAt      *time.Time
	GrantedBy      string
	IdempotencyKey string
}

// GrantCredit posts a promo credit grant journal and persists the grant record.
type GrantCredit struct {
	pool            *pgxpool.Pool
	credits         domaincredit.Repository
	ensureAccount   *EnsureCreditAccount
	postJournal     *ledgeruc.PostJournal
	audit           *audituc.Emitter
}

// NewGrantCredit creates the GrantCredit use case.
func NewGrantCredit(
	pool *pgxpool.Pool,
	credits domaincredit.Repository,
	ensureAccount *EnsureCreditAccount,
	postJournal *ledgeruc.PostJournal,
	audit *audituc.Emitter,
) *GrantCredit {
	return &GrantCredit{
		pool:          pool,
		credits:       credits,
		ensureAccount: ensureAccount,
		postJournal:   postJournal,
		audit:         audit,
	}
}

// Execute grants non-withdrawable promo credit to a user.
func (uc *GrantCredit) Execute(ctx context.Context, input GrantCreditInput) (*domaincredit.Grant, error) {
	if input.UserID == "" || input.Amount <= 0 || input.Reason == "" || input.GrantedBy == "" || input.IdempotencyKey == "" {
		return nil, domaincredit.ErrInvalidInput
	}

	grant := &domaincredit.Grant{
		UserID:         input.UserID,
		AmountRials:    input.Amount,
		Reason:         input.Reason,
		CampaignRef:    input.CampaignRef,
		ExpiresAt:      input.ExpiresAt,
		Status:         domaincredit.StatusActive,
		GrantedBy:      input.GrantedBy,
		IdempotencyKey: input.IdempotencyKey,
	}

	var result *domaincredit.Grant
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		if _, err := uc.ensureAccount.Execute(txCtx, input.UserID); err != nil {
			return err
		}
		if err := uc.credits.CreateGrant(txCtx, grant); err != nil {
			return err
		}

		_, err := uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypePromoCreditGrant,
			RefID:       grant.ID,
			Description: fmt.Sprintf("Promo credit grant: %s", input.Reason),
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.AccountAirbarPromoExpense, Debit: input.Amount, Credit: 0},
				{AccountCode: domainledger.UserPromoCreditAccount(input.UserID), Debit: 0, Credit: input.Amount},
			},
		})
		if err != nil {
			return err
		}

		_ = uc.audit.EmitCreditGranted(txCtx, grant.ID, grant.UserID, grant.AmountRials, string(grant.Status))
		result = grant
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
