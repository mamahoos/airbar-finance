package ledger

import "fmt"

const currencyIRT = "IRT"

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
)

// UserWalletAccount returns the wallet liability account for a user.
func UserWalletAccount(userID string) AccountCode {
	return AccountCode(fmt.Sprintf("USER:%s:%s:WALLET_LIABILITY", userID, currencyIRT))
}

// ShipmentEscrowAccount returns the escrow liability account for a shipment.
func ShipmentEscrowAccount(shipmentID string) AccountCode {
	return AccountCode(fmt.Sprintf("SHIPMENT:%s:%s:ESCROW", shipmentID, currencyIRT))
}
