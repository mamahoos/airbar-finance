package escrow

import (
	"context"
	"time"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

// CreateEscrowInput is the application input for UC-01.
type CreateEscrowInput struct {
	ShipmentID    string
	CarrierUserID string
	PayerUserID   string
	Amount        int64
}

// CreateEscrow creates a shipment escrow in CREATED status.
type CreateEscrow struct {
	repo   domainescrow.Repository
	audit  *audituc.Emitter
}

// NewCreateEscrow creates the CreateEscrow use case.
func NewCreateEscrow(repo domainescrow.Repository, audit *audituc.Emitter) *CreateEscrow {
	return &CreateEscrow{repo: repo, audit: audit}
}

// Execute validates input and persists a new escrow.
func (uc *CreateEscrow) Execute(ctx context.Context, input CreateEscrowInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" || input.CarrierUserID == "" || input.PayerUserID == "" {
		return nil, domainescrow.ErrInvalidAmount
	}
	if input.Amount <= 0 {
		return nil, domainescrow.ErrInvalidAmount
	}

	escrow := &domainescrow.Escrow{
		ShipmentID:    input.ShipmentID,
		CarrierUserID: input.CarrierUserID,
		PayerUserID:   input.PayerUserID,
		Amount:        input.Amount,
		Status:        domainescrow.StatusCreated,
	}

	if err := uc.repo.Create(ctx, escrow); err != nil {
		return nil, err
	}
	_ = uc.audit.EmitEscrowCreated(ctx, escrow.ID, escrow.ShipmentID, string(escrow.Status))
	return escrow, nil
}

// nowUTC returns the current UTC time (overridable in tests).
var nowUTC = func() time.Time { return time.Now().UTC() }
