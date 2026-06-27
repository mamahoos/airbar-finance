package wallet

import "time"

// Wallet is a read model for wallet balance and metadata (UC-14).
type Wallet struct {
	UserID      string
	Currency    string
	Balance     int64
	AccountCode string
}

// Transaction is a ledger-backed wallet movement (UC-15).
type Transaction struct {
	JournalID   string
	RefType     string
	RefID       string
	Description string
	Debit       int64
	Credit      int64
	CreatedAt   time.Time
}
