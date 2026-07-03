package escrow

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	credituc "github.com/mamahoos/airbar-finance/internal/usecase/credit"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

// PartialRefundEscrowInput is the application input for UC-08.
type PartialRefundEscrowInput struct {
	ShipmentID   string
	RefundAmount int64
}

// PartialRefundEscrow credits part of escrow to payer promo credit and/or wallet.
type PartialRefundEscrow struct {
	pool         *pgxpool.Pool
	escrowRepo   domainescrow.Repository
	postJournal  *ledgeruc.PostJournal
	ledger       LedgerBalanceReader
	ensureCredit *credituc.EnsureCreditAccount
	audit        *audituc.Emitter
}

// NewPartialRefundEscrow creates the PartialRefundEscrow use case.
func NewPartialRefundEscrow(
	pool *pgxpool.Pool,
	escrowRepo domainescrow.Repository,
	postJournal *ledgeruc.PostJournal,
	ledger LedgerBalanceReader,
	ensureCredit *credituc.EnsureCreditAccount,
	audit *audituc.Emitter,
) *PartialRefundEscrow {
	return &PartialRefundEscrow{
		pool:         pool,
		escrowRepo:   escrowRepo,
		postJournal:  postJournal,
		ledger:       ledger,
		ensureCredit: ensureCredit,
		audit:        audit,
	}
}

// Execute posts partial refund journals.
func (uc *PartialRefundEscrow) Execute(ctx context.Context, input PartialRefundEscrowInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" || input.RefundAmount <= 0 {
		return nil, domainescrow.ErrInvalidAmount
	}

	var result *domainescrow.Escrow
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		escrow, err := uc.escrowRepo.GetByShipmentID(txCtx, input.ShipmentID)
		if err != nil {
			return err
		}
		if !escrow.Status.CanPartialRefund() {
			return domainescrow.ErrInvalidTransition
		}

		balance, err := EscrowBalance(txCtx, uc.ledger, input.ShipmentID)
		if err != nil {
			return err
		}
		if balance <= 0 {
			return domainescrow.ErrNoEscrowBalance
		}
		if input.RefundAmount > balance {
			return domainescrow.ErrRefundExceedsBalance
		}

		promoRefund, walletRefund := RefundAllocation(input.RefundAmount, escrow.PromoCreditFunded)
		if err := ensureCreditAccountIfNeeded(txCtx, uc.ensureCredit, escrow.PayerUserID, promoRefund); err != nil {
			return err
		}
		if err := postEscrowRefundJournals(txCtx, uc.postJournal, escrow.PayerUserID, input.ShipmentID, promoRefund, walletRefund); err != nil {
			return err
		}

		now := nowUTC()
		escrow.PromoCreditFunded -= promoRefund
		if escrow.PromoCreditFunded < 0 {
			escrow.PromoCreditFunded = 0
		}
		if input.RefundAmount == balance {
			escrow.Status = domainescrow.StatusRefunded
		} else {
			escrow.Status = domainescrow.StatusPartiallyRefunded
		}
		escrow.RefundedAt = &now
		if err := uc.escrowRepo.Update(txCtx, escrow); err != nil {
			return err
		}
		emitEscrowStatus(txCtx, uc.audit, escrow)
		result = escrow
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
