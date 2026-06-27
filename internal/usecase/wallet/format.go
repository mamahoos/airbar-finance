package wallet

import (
	"strconv"

	domainwallet "github.com/mamahoos/airbar-finance/internal/domain/wallet"
)

// FormatAmount formats rials as a decimal string for proto responses.
func FormatAmount(amount int64) string {
	return strconv.FormatInt(amount, 10)
}

// NormalizeCurrency defaults empty currency to IRT and validates support.
func NormalizeCurrency(currency string) (string, error) {
	if currency == "" {
		return domainwallet.CurrencyIRT, nil
	}
	if currency != domainwallet.CurrencyIRT {
		return "", domainwallet.ErrUnsupportedCurrency
	}
	return currency, nil
}
