package provider

import "time"

const ProviderZibal = "ZIBAL"

// EventType classifies a provider interaction.
type EventType string

const (
	EventTypeRequest  EventType = "REQUEST"
	EventTypeVerify   EventType = "VERIFY"
	EventTypeCallback EventType = "CALLBACK"
)

// Event is an audit record for a PSP interaction.
type Event struct {
	ID             string
	Provider       string
	EventType      EventType
	PaymentOrderID string
	Payload        []byte
	PayloadHash    string
	IdempotencyKey string
	Processed      bool
	CreatedAt      time.Time
}
