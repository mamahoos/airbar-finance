package withdrawal

import (
	"context"

	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

// ProcessWithdrawalInput is the application input for UC-18.
type ProcessWithdrawalInput struct {
	WithdrawalID string
	ProviderRef  string
}

// ProcessWithdrawal marks a pending withdrawal as COMPLETED.
type ProcessWithdrawal struct {
	withdrawals domainwithdrawal.Repository
	audit       *audituc.Emitter
}

// NewProcessWithdrawal creates the ProcessWithdrawal use case.
func NewProcessWithdrawal(withdrawals domainwithdrawal.Repository, audit *audituc.Emitter) *ProcessWithdrawal {
	return &ProcessWithdrawal{withdrawals: withdrawals, audit: audit}
}

// Execute completes a reserved withdrawal after admin payout.
func (uc *ProcessWithdrawal) Execute(ctx context.Context, input ProcessWithdrawalInput) (*domainwithdrawal.Withdrawal, error) {
	if input.WithdrawalID == "" {
		return nil, domainwithdrawal.ErrInvalidInput
	}

	withdrawal, err := uc.withdrawals.GetByID(ctx, input.WithdrawalID)
	if err != nil {
		return nil, err
	}
	if withdrawal.Status != domainwithdrawal.StatusPending {
		return nil, domainwithdrawal.ErrInvalidTransition
	}

	now := nowUTC()
	withdrawal.Status = domainwithdrawal.StatusCompleted
	withdrawal.ProviderRef = input.ProviderRef
	withdrawal.ProcessedAt = &now
	if err := uc.withdrawals.Update(ctx, withdrawal); err != nil {
		return nil, err
	}
	_ = uc.audit.EmitWithdrawalStatusChanged(ctx, withdrawal.ID, string(withdrawal.Status))
	return withdrawal, nil
}
