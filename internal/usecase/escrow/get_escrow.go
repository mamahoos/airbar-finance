package escrow

import (
	"context"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
)

// GetEscrow loads an escrow by shipment id (UC-02).
type GetEscrow struct {
	repo domainescrow.Repository
}

// NewGetEscrow creates the GetEscrow use case.
func NewGetEscrow(repo domainescrow.Repository) *GetEscrow {
	return &GetEscrow{repo: repo}
}

// Execute returns the escrow aggregate.
func (uc *GetEscrow) Execute(ctx context.Context, shipmentID string) (*domainescrow.Escrow, error) {
	return uc.repo.GetByShipmentID(ctx, shipmentID)
}
