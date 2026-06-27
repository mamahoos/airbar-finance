package withdrawal

import "time"

// Status is the withdrawal lifecycle state.
type Status string

const (
	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusRejected  Status = "REJECTED"
)

// Withdrawal is a payout request aggregate.
type Withdrawal struct {
	ID              string
	UserID          string
	Amount          int64
	Status          Status
	DestinationHash string
	ProviderRef     string
	RejectReason    string
	ProcessedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
