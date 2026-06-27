package escrow

import (
	"context"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
)

// FreezeEscrowInput is the application input for UC-05.
type FreezeEscrowInput struct {
	ShipmentID string
}

// FreezeEscrow freezes an escrow during dispute.
type FreezeEscrow struct {
	repo domainescrow.Repository
}

// NewFreezeEscrow creates the FreezeEscrow use case.
func NewFreezeEscrow(repo domainescrow.Repository) *FreezeEscrow {
	return &FreezeEscrow{repo: repo}
}

// Execute transitions escrow to FROZEN.
func (uc *FreezeEscrow) Execute(ctx context.Context, input FreezeEscrowInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" {
		return nil, domainescrow.ErrInvalidAmount
	}

	escrow, err := uc.repo.GetByShipmentID(ctx, input.ShipmentID)
	if err != nil {
		return nil, err
	}
	if !escrow.Status.CanFreeze() {
		return nil, domainescrow.ErrInvalidTransition
	}

	escrow.Status = domainescrow.StatusFrozen
	if err := uc.repo.Update(ctx, escrow); err != nil {
		return nil, err
	}
	return escrow, nil
}
