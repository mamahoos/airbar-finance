package credit

import (
	"context"

	domaincredit "github.com/mamahoos/airbar-finance/internal/domain/credit"
)

// ListGrants returns promo credit grants for a user with current balance.
type ListGrants struct {
	credits    domaincredit.Repository
	getBalance *GetBalance
}

// NewListGrants creates the ListGrants use case.
func NewListGrants(credits domaincredit.Repository, getBalance *GetBalance) *ListGrants {
	return &ListGrants{credits: credits, getBalance: getBalance}
}

// ListGrantsResult is the read model for admin/user credit listing.
type ListGrantsResult struct {
	Balance int64
	Grants  []domaincredit.Grant
}

// Execute lists grants and returns the current promo credit balance.
func (uc *ListGrants) Execute(ctx context.Context, userID string, limit, offset int) (*ListGrantsResult, error) {
	if userID == "" {
		return nil, domaincredit.ErrInvalidInput
	}

	grants, err := uc.credits.ListGrantsByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	balance, err := uc.getBalance.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ListGrantsResult{Balance: balance, Grants: grants}, nil
}
