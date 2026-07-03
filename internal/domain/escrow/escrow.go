package escrow

import "time"

// Status is the escrow lifecycle state.
type Status string

const (
	StatusCreated           Status = "CREATED"
	StatusFunded            Status = "FUNDED"
	StatusLocked            Status = "LOCKED"
	StatusDisputeWindow     Status = "DISPUTE_WINDOW"
	StatusFrozen            Status = "FROZEN"
	StatusReleased          Status = "RELEASED"
	StatusRefunded          Status = "REFUNDED"
	StatusPartiallyRefunded Status = "PARTIALLY_REFUNDED"
)

// FundingSource records how an escrow was funded.
type FundingSource string

const (
	FundingSourcePSP         FundingSource = "PSP"
	FundingSourceWallet      FundingSource = "WALLET"
	FundingSourcePromoCredit FundingSource = "PROMO_CREDIT"
	FundingSourceMixed       FundingSource = "MIXED"
)

// Escrow is the shipment escrow aggregate root.
type Escrow struct {
	ID                string
	ShipmentID        string
	CarrierUserID     string
	PayerUserID       string
	Amount            int64
	Status            Status
	PaymentOrderID    string
	FundingSource     FundingSource
	PromoCreditFunded int64
	FundedAt          *time.Time
	ReleasedAt        *time.Time
	RefundedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
