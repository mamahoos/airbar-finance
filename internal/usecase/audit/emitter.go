package audit

import (
	"context"

	domainaudit "github.com/mamahoos/airbar-finance/internal/domain/audit"
)

// Emitter records finance audit events.
type Emitter struct {
	repo domainaudit.Repository
}

// NewEmitter creates an audit emitter.
func NewEmitter(repo domainaudit.Repository) *Emitter {
	return &Emitter{repo: repo}
}

// Emit persists an audit event (no-op when repo is nil).
func (e *Emitter) Emit(ctx context.Context, aggregateType, aggregateID string, eventType domainaudit.EventType, payload map[string]any) error {
	if e == nil || e.repo == nil {
		return nil
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return e.repo.Create(ctx, &domainaudit.Event{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       payload,
	})
}

// EmitEscrowCreated records escrow creation.
func (e *Emitter) EmitEscrowCreated(ctx context.Context, escrowID, shipmentID, status string) error {
	return e.Emit(ctx, domainaudit.AggregateEscrow, escrowID, domainaudit.EventEscrowCreated, map[string]any{
		"shipment_id": shipmentID,
		"status":      status,
	})
}

// EmitEscrowStatusChanged records an escrow transition.
func (e *Emitter) EmitEscrowStatusChanged(ctx context.Context, escrowID, shipmentID, status string) error {
	return e.Emit(ctx, domainaudit.AggregateEscrow, escrowID, domainaudit.EventEscrowStatusChanged, map[string]any{
		"shipment_id": shipmentID,
		"status":      status,
	})
}

// EmitPaymentCreated records payment order creation.
func (e *Emitter) EmitPaymentCreated(ctx context.Context, orderID, purpose, status string) error {
	return e.Emit(ctx, domainaudit.AggregatePaymentOrder, orderID, domainaudit.EventPaymentCreated, map[string]any{
		"purpose": purpose,
		"status":  status,
	})
}

// EmitPaymentStatusChanged records payment order status change.
func (e *Emitter) EmitPaymentStatusChanged(ctx context.Context, orderID, status string) error {
	return e.Emit(ctx, domainaudit.AggregatePaymentOrder, orderID, domainaudit.EventPaymentStatusChanged, map[string]any{
		"status": status,
	})
}

// EmitWithdrawalCreated records withdrawal creation.
func (e *Emitter) EmitWithdrawalCreated(ctx context.Context, withdrawalID, userID, status string) error {
	return e.Emit(ctx, domainaudit.AggregateWithdrawal, withdrawalID, domainaudit.EventWithdrawalCreated, map[string]any{
		"user_id": userID,
		"status":  status,
	})
}

// EmitWithdrawalStatusChanged records withdrawal status change.
func (e *Emitter) EmitWithdrawalStatusChanged(ctx context.Context, withdrawalID, status string) error {
	return e.Emit(ctx, domainaudit.AggregateWithdrawal, withdrawalID, domainaudit.EventWithdrawalStatusChanged, map[string]any{
		"status": status,
	})
}

// EmitCreditGranted records promo credit grant.
func (e *Emitter) EmitCreditGranted(ctx context.Context, grantID, userID string, amount int64, status string) error {
	return e.Emit(ctx, domainaudit.AggregateCreditGrant, grantID, domainaudit.EventCreditGranted, map[string]any{
		"user_id": userID,
		"amount":  amount,
		"status":  status,
	})
}

// EmitCreditReversed records promo credit reversal.
func (e *Emitter) EmitCreditReversed(ctx context.Context, grantID, userID string, amount int64, reason string) error {
	return e.Emit(ctx, domainaudit.AggregateCreditGrant, grantID, domainaudit.EventCreditReversed, map[string]any{
		"user_id": userID,
		"amount":  amount,
		"reason":  reason,
	})
}
