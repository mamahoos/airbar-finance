package audit

import "time"

const (
	AggregateEscrow       = "escrow"
	AggregatePaymentOrder = "payment_order"
	AggregateWithdrawal   = "withdrawal"
)

// EventType classifies finance audit events.
type EventType string

const (
	EventEscrowCreated       EventType = "ESCROW_CREATED"
	EventEscrowStatusChanged EventType = "ESCROW_STATUS_CHANGED"
	EventPaymentCreated      EventType = "PAYMENT_CREATED"
	EventPaymentStatusChanged EventType = "PAYMENT_STATUS_CHANGED"
	EventWithdrawalCreated   EventType = "WITHDRAWAL_CREATED"
	EventWithdrawalStatusChanged EventType = "WITHDRAWAL_STATUS_CHANGED"
)

// Event is an immutable finance audit record.
type Event struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     EventType
	Payload       map[string]any
	CreatedAt     time.Time
}
