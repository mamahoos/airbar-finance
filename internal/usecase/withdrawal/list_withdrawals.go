package withdrawal

import (
	"context"

	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
)

// ListWithdrawals lists withdrawals for a user (UC-17).
type ListWithdrawals struct {
	withdrawals domainwithdrawal.Repository
}

// NewListWithdrawals creates the ListWithdrawals use case.
func NewListWithdrawals(withdrawals domainwithdrawal.Repository) *ListWithdrawals {
	return &ListWithdrawals{withdrawals: withdrawals}
}

// Execute returns withdrawals filtered by user and optional status.
func (uc *ListWithdrawals) Execute(ctx context.Context, userID, status string) ([]domainwithdrawal.Withdrawal, error) {
	if userID == "" {
		return nil, domainwithdrawal.ErrInvalidInput
	}
	return uc.withdrawals.List(ctx, userID, domainwithdrawal.Status(status))
}
