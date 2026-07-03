package withdrawal

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
)

func emitWithdrawalStatusChanged(ctx context.Context, audit *audituc.Emitter, withdrawalID string, status domainwithdrawal.Status) {
	if audit == nil {
		return
	}
	_ = audit.EmitWithdrawalStatusChanged(ctx, withdrawalID, string(status))
}

// ApproveWithdrawal moves a payout request from PENDING to APPROVED.
type ApproveWithdrawal struct {
	withdrawals domainwithdrawal.Repository
	audit       *audituc.Emitter
}

func NewApproveWithdrawal(withdrawals domainwithdrawal.Repository, audit *audituc.Emitter) *ApproveWithdrawal {
	return &ApproveWithdrawal{withdrawals: withdrawals, audit: audit}
}

func (uc *ApproveWithdrawal) Execute(ctx context.Context, withdrawalID string) (*domainwithdrawal.Withdrawal, error) {
	if withdrawalID == "" {
		return nil, domainwithdrawal.ErrInvalidInput
	}
	withdrawal, err := uc.withdrawals.GetByID(ctx, withdrawalID)
	if err != nil {
		return nil, err
	}
	if withdrawal.Status != domainwithdrawal.StatusPending {
		return nil, domainwithdrawal.ErrInvalidTransition
	}
	withdrawal.Status = domainwithdrawal.StatusApproved
	if err := uc.withdrawals.Update(ctx, withdrawal); err != nil {
		return nil, err
	}
	emitWithdrawalStatusChanged(ctx, uc.audit, withdrawal.ID, withdrawal.Status)
	return withdrawal, nil
}

type MarkWithdrawalSentInput struct {
	WithdrawalID  string
	ProviderRef   string
	PayoutChannel string
	ReceiptURL    string
}

// MarkWithdrawalSent records provider dispatch details.
type MarkWithdrawalSent struct {
	withdrawals domainwithdrawal.Repository
	audit       *audituc.Emitter
}

func NewMarkWithdrawalSent(withdrawals domainwithdrawal.Repository, audit *audituc.Emitter) *MarkWithdrawalSent {
	return &MarkWithdrawalSent{withdrawals: withdrawals, audit: audit}
}

func (uc *MarkWithdrawalSent) Execute(ctx context.Context, input MarkWithdrawalSentInput) (*domainwithdrawal.Withdrawal, error) {
	if input.WithdrawalID == "" || input.ProviderRef == "" || input.PayoutChannel == "" || input.ReceiptURL == "" {
		return nil, domainwithdrawal.ErrInvalidInput
	}
	withdrawal, err := uc.withdrawals.GetByID(ctx, input.WithdrawalID)
	if err != nil {
		return nil, err
	}
	if withdrawal.Status != domainwithdrawal.StatusApproved {
		return nil, domainwithdrawal.ErrInvalidTransition
	}
	withdrawal.Status = domainwithdrawal.StatusSentToBank
	withdrawal.ProviderRef = input.ProviderRef
	withdrawal.PayoutChannel = input.PayoutChannel
	withdrawal.ReceiptURL = input.ReceiptURL
	if err := uc.withdrawals.Update(ctx, withdrawal); err != nil {
		return nil, err
	}
	emitWithdrawalStatusChanged(ctx, uc.audit, withdrawal.ID, withdrawal.Status)
	return withdrawal, nil
}

// SettleWithdrawal marks a bank-sent payout as settled.
type SettleWithdrawal struct {
	withdrawals domainwithdrawal.Repository
	audit       *audituc.Emitter
}

func NewSettleWithdrawal(withdrawals domainwithdrawal.Repository, audit *audituc.Emitter) *SettleWithdrawal {
	return &SettleWithdrawal{withdrawals: withdrawals, audit: audit}
}

func (uc *SettleWithdrawal) Execute(ctx context.Context, withdrawalID string) (*domainwithdrawal.Withdrawal, error) {
	if withdrawalID == "" {
		return nil, domainwithdrawal.ErrInvalidInput
	}
	withdrawal, err := uc.withdrawals.GetByID(ctx, withdrawalID)
	if err != nil {
		return nil, err
	}
	if withdrawal.Status != domainwithdrawal.StatusSentToBank {
		return nil, domainwithdrawal.ErrInvalidTransition
	}
	now := nowUTC()
	withdrawal.Status = domainwithdrawal.StatusSettled
	withdrawal.ProcessedAt = &now
	if err := uc.withdrawals.Update(ctx, withdrawal); err != nil {
		return nil, err
	}
	emitWithdrawalStatusChanged(ctx, uc.audit, withdrawal.ID, withdrawal.Status)
	return withdrawal, nil
}

type FailWithdrawalInput struct {
	WithdrawalID string
	Reason       string
}

// FailWithdrawal reverses reserved funds after a payout provider failure.
type FailWithdrawal struct {
	pool        *pgxpool.Pool
	withdrawals domainwithdrawal.Repository
	postJournal *ledgeruc.PostJournal
	audit       *audituc.Emitter
}

func NewFailWithdrawal(
	pool *pgxpool.Pool,
	withdrawals domainwithdrawal.Repository,
	postJournal *ledgeruc.PostJournal,
	audit *audituc.Emitter,
) *FailWithdrawal {
	return &FailWithdrawal{pool: pool, withdrawals: withdrawals, postJournal: postJournal, audit: audit}
}

func (uc *FailWithdrawal) Execute(ctx context.Context, input FailWithdrawalInput) (*domainwithdrawal.Withdrawal, error) {
	if input.WithdrawalID == "" || input.Reason == "" {
		return nil, domainwithdrawal.ErrInvalidInput
	}

	var result *domainwithdrawal.Withdrawal
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		withdrawal, err := uc.withdrawals.GetByID(txCtx, input.WithdrawalID)
		if err != nil {
			return err
		}
		if withdrawal.Status != domainwithdrawal.StatusApproved && withdrawal.Status != domainwithdrawal.StatusSentToBank {
			return domainwithdrawal.ErrInvalidTransition
		}

		_, err = uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypeWithdrawalRejectReversal,
			RefID:       fmt.Sprintf("%s:fail", withdrawal.ID),
			Description: "Withdrawal failure reversal",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.AccountIRPayoutClearing, Debit: withdrawal.Amount, Credit: 0},
				{AccountCode: domainledger.UserWalletAccount(withdrawal.UserID), Debit: 0, Credit: withdrawal.Amount},
			},
		})
		if err != nil {
			return err
		}

		now := nowUTC()
		withdrawal.Status = domainwithdrawal.StatusFailed
		withdrawal.RejectReason = input.Reason
		withdrawal.ProcessedAt = &now
		if err := uc.withdrawals.Update(txCtx, withdrawal); err != nil {
			return err
		}
		emitWithdrawalStatusChanged(txCtx, uc.audit, withdrawal.ID, withdrawal.Status)
		result = withdrawal
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
