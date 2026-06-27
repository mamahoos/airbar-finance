package payment

import "time"

// Purpose classifies why a payment order exists.
type Purpose string

const (
	PurposeShipment    Purpose = "SHIPMENT"
	PurposeWalletTopup Purpose = "WALLET_TOPUP"
)

// Status is the payment order lifecycle state.
type Status string

const (
	StatusPending   Status = "PENDING"
	StatusConfirmed Status = "CONFIRMED"
	StatusFailed    Status = "FAILED"
)

// Order is a Zibal payment order aggregate.
type Order struct {
	ID          string
	ShipmentID  string
	PayerUserID string
	Purpose     Purpose
	Amount      int64
	Status      Status
	Authority   string
	RedirectURL string
	SuccessURL  string
	FailureURL  string
	Description string
	AgreedPrice int64
	VerifiedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
