package escrow

import (
	"context"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
)

// MarkDeliveredInput is the application input for UC-04.
type MarkDeliveredInput struct {
	ShipmentID string
}

// MarkDelivered transitions a funded escrow into DISPUTE_WINDOW.
type MarkDelivered struct {
	repo domainescrow.Repository
}

// NewMarkDelivered creates the MarkDelivered use case.
func NewMarkDelivered(repo domainescrow.Repository) *MarkDelivered {
	return &MarkDelivered{repo: repo}
}

// Execute updates escrow status after carrier delivery.
func (uc *MarkDelivered) Execute(ctx context.Context, input MarkDeliveredInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" {
		return nil, domainescrow.ErrInvalidAmount
	}

	escrow, err := uc.repo.GetByShipmentID(ctx, input.ShipmentID)
	if err != nil {
		return nil, err
	}
	if !escrow.Status.CanMarkDelivered() {
		return nil, domainescrow.ErrInvalidTransition
	}

	escrow.Status = domainescrow.StatusDisputeWindow
	if err := uc.repo.Update(ctx, escrow); err != nil {
		return nil, err
	}
	return escrow, nil
}
