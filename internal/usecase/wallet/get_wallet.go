package wallet

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainwallet "github.com/mamahoos/airbar-finance/internal/domain/wallet"
)

// GetWallet returns wallet metadata and ledger-derived balance (UC-14).
type GetWallet struct {
	getBalance *GetBalance
}

// NewGetWallet creates the GetWallet use case.
func NewGetWallet(getBalance *GetBalance) *GetWallet {
	return &GetWallet{getBalance: getBalance}
}

// Execute returns wallet read model for a user.
func (uc *GetWallet) Execute(ctx context.Context, userID, currency string) (*domainwallet.Wallet, error) {
	if userID == "" {
		return nil, domainwallet.ErrInvalidInput
	}

	normalizedCurrency, err := NormalizeCurrency(currency)
	if err != nil {
		return nil, err
	}

	balance, err := uc.getBalance.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domainwallet.Wallet{
		UserID:      userID,
		Currency:    normalizedCurrency,
		Balance:     balance,
		AccountCode: domainledger.UserWalletAccount(userID).String(),
	}, nil
}
