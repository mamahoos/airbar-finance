package idempotency

import "time"

// Status tracks idempotency record lifecycle.
type Status string

const (
	StatusProcessing Status = "PROCESSING"
	StatusCompleted  Status = "COMPLETED"
)

// Record stores dedup state for a mutating command.
type Record struct {
	Key              string
	Scope            string
	ResourceType     string
	ResourceID       string
	Status           Status
	ResponseSnapshot map[string]any
	CreatedAt        time.Time
	CompletedAt      *time.Time
}
