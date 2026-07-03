package treasury

import (
	"context"

	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
)

const defaultCurrency = "IRT"

// Summary is treasury account balances for ops.
type Summary struct {
	Currency string
	Accounts map[string]int64
}

// LedgerReader supports treasury balance queries.
type LedgerReader interface {
	SumByAccount(ctx context.Context, accountCode domainledger.AccountCode) (debit int64, credit int64, err error)
	SumByAccountLike(ctx context.Context, pattern string) (debit int64, credit int64, err error)
}

// GetTreasurySummary returns system and aggregate ledger balances (UC-20).
type GetTreasurySummary struct {
	ledger LedgerReader
}

// NewGetTreasurySummary creates the GetTreasurySummary use case.
func NewGetTreasurySummary(ledger LedgerReader) *GetTreasurySummary {
	return &GetTreasurySummary{ledger: ledger}
}

// Execute returns treasury balances in rials. Only IRT is supported today.
func (uc *GetTreasurySummary) Execute(ctx context.Context, currency string) (*Summary, error) {
	if currency == "" {
		currency = defaultCurrency
	}
	if currency != defaultCurrency {
		return nil, ErrUnsupportedCurrency
	}

	accounts := make(map[string]int64)

	psp, err := uc.assetBalance(ctx, domainledger.AccountIRPSPClearing)
	if err != nil {
		return nil, err
	}
	accounts[string(domainledger.AccountIRPSPClearing)] = psp

	bank, err := uc.assetBalance(ctx, domainledger.AccountIRBankMain)
	if err != nil {
		return nil, err
	}
	accounts[string(domainledger.AccountIRBankMain)] = bank

	payout, err := uc.assetBalance(ctx, domainledger.AccountIRPayoutClearing)
	if err != nil {
		return nil, err
	}
	accounts[string(domainledger.AccountIRPayoutClearing)] = payout

	fee, err := uc.revenueBalance(ctx, domainledger.AccountAirbarFeeRevenue)
	if err != nil {
		return nil, err
	}
	accounts[string(domainledger.AccountAirbarFeeRevenue)] = fee

	promoExpense, err := uc.expenseBalance(ctx, domainledger.AccountAirbarPromoExpense)
	if err != nil {
		return nil, err
	}
	accounts[string(domainledger.AccountAirbarPromoExpense)] = promoExpense

	walletNet, err := uc.liabilityBalance(ctx, domainledger.WalletAccountLikePattern())
	if err != nil {
		return nil, err
	}
	accounts["AGGREGATE_WALLET_LIABILITY"] = walletNet

	escrowNet, err := uc.liabilityBalance(ctx, domainledger.EscrowAccountLikePattern())
	if err != nil {
		return nil, err
	}
	accounts["AGGREGATE_ESCROW_LIABILITY"] = escrowNet

	promoNet, err := uc.liabilityBalance(ctx, domainledger.PromoCreditAccountLikePattern())
	if err != nil {
		return nil, err
	}
	accounts["AGGREGATE_PROMO_CREDIT_LIABILITY"] = promoNet

	return &Summary{Currency: currency, Accounts: accounts}, nil
}

func (uc *GetTreasurySummary) assetBalance(ctx context.Context, code domainledger.AccountCode) (int64, error) {
	debit, credit, err := uc.ledger.SumByAccount(ctx, code)
	if err != nil {
		return 0, err
	}
	return debit - credit, nil
}

func (uc *GetTreasurySummary) revenueBalance(ctx context.Context, code domainledger.AccountCode) (int64, error) {
	debit, credit, err := uc.ledger.SumByAccount(ctx, code)
	if err != nil {
		return 0, err
	}
	return credit - debit, nil
}

func (uc *GetTreasurySummary) expenseBalance(ctx context.Context, code domainledger.AccountCode) (int64, error) {
	debit, credit, err := uc.ledger.SumByAccount(ctx, code)
	if err != nil {
		return 0, err
	}
	return debit - credit, nil
}

func (uc *GetTreasurySummary) liabilityBalance(ctx context.Context, pattern string) (int64, error) {
	debit, credit, err := uc.ledger.SumByAccountLike(ctx, pattern)
	if err != nil {
		return 0, err
	}
	return credit - debit, nil
}
