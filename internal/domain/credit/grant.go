package credit

import "time"

// GrantStatus tracks promo credit grant lifecycle.
type GrantStatus string

const (
	StatusActive   GrantStatus = "ACTIVE"
	StatusReversed GrantStatus = "REVERSED"
)

// Grant is a persisted promo credit grant record.
type Grant struct {
	ID             string
	UserID         string
	AmountRials    int64
	Reason         string
	CampaignRef    string
	ExpiresAt      *time.Time
	Status         GrantStatus
	GrantedBy      string
	IdempotencyKey string
	ReversedAt     *time.Time
	ReverseReason  string
	ReversedBy     string
	CreatedAt      time.Time
}
