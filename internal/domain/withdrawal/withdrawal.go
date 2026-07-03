package withdrawal

import "time"

// Status is the withdrawal lifecycle state.
type Status string

const (
	StatusPending    Status = "PENDING"
	StatusApproved   Status = "APPROVED"
	StatusSentToBank Status = "SENT_TO_BANK"
	StatusSettled    Status = "SETTLED"
	StatusFailed     Status = "FAILED"
	StatusCompleted  Status = "COMPLETED"
	StatusRejected   Status = "REJECTED"
)

// Withdrawal is a payout request aggregate.
type Withdrawal struct {
	ID              string
	UserID          string
	Amount          int64
	Status          Status
	DestinationHash string
	ProviderRef     string
	PayoutChannel   string
	ReceiptURL      string
	RejectReason    string
	ProcessedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
