package withdrawal

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
)

// RejectWithdrawalInput is the application input for UC-19.
type RejectWithdrawalInput struct {
	WithdrawalID string
	Reason       string
}

// RejectWithdrawal reverses a reserved withdrawal back to the user wallet.
type RejectWithdrawal struct {
	pool        *pgxpool.Pool
	withdrawals domainwithdrawal.Repository
	postJournal *ledgeruc.PostJournal
}

// NewRejectWithdrawal creates the RejectWithdrawal use case.
func NewRejectWithdrawal(
	pool *pgxpool.Pool,
	withdrawals domainwithdrawal.Repository,
	postJournal *ledgeruc.PostJournal,
) *RejectWithdrawal {
	return &RejectWithdrawal{
		pool:        pool,
		withdrawals: withdrawals,
		postJournal: postJournal,
	}
}

// Execute posts WITHDRAWAL_REJECT_REVERSAL and marks withdrawal REJECTED.
func (uc *RejectWithdrawal) Execute(ctx context.Context, input RejectWithdrawalInput) (*domainwithdrawal.Withdrawal, error) {
	if input.WithdrawalID == "" {
		return nil, domainwithdrawal.ErrInvalidInput
	}

	var result *domainwithdrawal.Withdrawal
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		withdrawal, err := uc.withdrawals.GetByID(txCtx, input.WithdrawalID)
		if err != nil {
			return err
		}
		if withdrawal.Status != domainwithdrawal.StatusPending {
			return domainwithdrawal.ErrInvalidTransition
		}

		_, err = uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypeWithdrawalRejectReversal,
			RefID:       fmt.Sprintf("%s:reject", withdrawal.ID),
			Description: "Withdrawal reject reversal",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.AccountIRPayoutClearing, Debit: withdrawal.Amount, Credit: 0},
				{AccountCode: domainledger.UserWalletAccount(withdrawal.UserID), Debit: 0, Credit: withdrawal.Amount},
			},
		})
		if err != nil {
			return err
		}

		now := nowUTC()
		withdrawal.Status = domainwithdrawal.StatusRejected
		withdrawal.RejectReason = input.Reason
		withdrawal.ProcessedAt = &now
		if err := uc.withdrawals.Update(txCtx, withdrawal); err != nil {
			return err
		}
		result = withdrawal
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
