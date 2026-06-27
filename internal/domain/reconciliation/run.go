package reconciliation

import "time"

// Status is the outcome of a reconciliation run.
type Status string

const (
	StatusPassed Status = "PASSED"
	StatusFailed Status = "FAILED"
)

// Run records a ledger integrity check.
type Run struct {
	ID          string
	Status      Status
	Findings    map[string]any
	StartedAt   time.Time
	CompletedAt *time.Time
}
