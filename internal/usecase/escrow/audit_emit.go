package escrow

import (
	"context"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

func emitEscrowStatus(ctx context.Context, audit *audituc.Emitter, escrow *domainescrow.Escrow) {
	if escrow == nil {
		return
	}
	_ = audit.EmitEscrowStatusChanged(ctx, escrow.ID, escrow.ShipmentID, string(escrow.Status))
}
