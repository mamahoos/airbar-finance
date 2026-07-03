package credit

import "time"

const CurrencyIRT = "IRT"

// Account is a registered promo credit account for a user. Balance is never stored here.
type Account struct {
	ID          string
	UserID      string
	Currency    string
	AccountCode string
	CreatedAt   time.Time
}
