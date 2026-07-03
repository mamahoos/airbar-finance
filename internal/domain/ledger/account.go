package ledger

import "fmt"

const currencyIRT = "IRT"

const (
	walletAccountPrefix      = "USER:"
	walletAccountSuffix      = ":IRT:WALLET_LIABILITY"
	promoCreditAccountSuffix = ":IRT:PROMO_CREDIT_LIABILITY"
	escrowAccountSuffix      = ":IRT:ESCROW"
)

// WalletAccountLikePattern matches all user wallet liability accounts.
func WalletAccountLikePattern() string {
	return walletAccountPrefix + "%" + walletAccountSuffix
}

// EscrowAccountLikePattern matches all shipment escrow accounts.
func EscrowAccountLikePattern() string {
	return "SHIPMENT:%" + escrowAccountSuffix
}

// PromoCreditAccountLikePattern matches all user promo credit liability accounts.
func PromoCreditAccountLikePattern() string {
	return walletAccountPrefix + "%" + promoCreditAccountSuffix
}

// AccountCode identifies a ledger account (SSOT for balances via SUM(entries)).
type AccountCode string

func (c AccountCode) String() string {
	return string(c)
}

// System account codes (constants — not DB seed rows).
const (
	AccountIRPSPClearing    AccountCode = "IR_PSP_CLEARING"
	AccountIRBankMain       AccountCode = "IR_BANK_MAIN"
	AccountIRPayoutClearing AccountCode = "IR_PAYOUT_CLEARING"
	AccountAirbarFeeRevenue AccountCode = "AIRBAR_FEE_REVENUE"
	AccountAirbarPromoExpense AccountCode = "AIRBAR_PROMO_EXPENSE"
)

// UserWalletAccount returns the wallet liability account for a user.
func UserWalletAccount(userID string) AccountCode {
	return AccountCode(fmt.Sprintf("USER:%s:%s:WALLET_LIABILITY", userID, currencyIRT))
}

// UserPromoCreditAccount returns the non-withdrawable promo credit liability account for a user.
func UserPromoCreditAccount(userID string) AccountCode {
	return AccountCode(fmt.Sprintf("USER:%s:%s:PROMO_CREDIT_LIABILITY", userID, currencyIRT))
}

// ShipmentEscrowAccount returns the escrow liability account for a shipment.
func ShipmentEscrowAccount(shipmentID string) AccountCode {
	return AccountCode(fmt.Sprintf("SHIPMENT:%s:%s:ESCROW", shipmentID, currencyIRT))
}

// ParseUserIDFromWalletAccount extracts user_id from a wallet liability account code.
func ParseUserIDFromWalletAccount(code AccountCode) (string, bool) {
	raw := code.String()
	if len(raw) <= len(walletAccountPrefix)+len(walletAccountSuffix) {
		return "", false
	}
	if raw[:len(walletAccountPrefix)] != walletAccountPrefix {
		return "", false
	}
	if raw[len(raw)-len(walletAccountSuffix):] != walletAccountSuffix {
		return "", false
	}
	userID := raw[len(walletAccountPrefix) : len(raw)-len(walletAccountSuffix)]
	if userID == "" {
		return "", false
	}
	return userID, true
}

// ParseUserIDFromPromoCreditAccount extracts user_id from a promo credit liability account code.
func ParseUserIDFromPromoCreditAccount(code AccountCode) (string, bool) {
	raw := code.String()
	if len(raw) <= len(walletAccountPrefix)+len(promoCreditAccountSuffix) {
		return "", false
	}
	if raw[:len(walletAccountPrefix)] != walletAccountPrefix {
		return "", false
	}
	if raw[len(raw)-len(promoCreditAccountSuffix):] != promoCreditAccountSuffix {
		return "", false
	}
	userID := raw[len(walletAccountPrefix) : len(raw)-len(promoCreditAccountSuffix)]
	if userID == "" {
		return "", false
	}
	return userID, true
}
