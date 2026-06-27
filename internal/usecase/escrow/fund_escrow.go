package escrow

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
)

// FundEscrowInput is the application input for internal PSP funding.
type FundEscrowInput struct {
	ShipmentID     string
	PaymentOrderID string
}

// FundEscrow marks an escrow FUNDED after PSP payment verification.
type FundEscrow struct {
	pool        *pgxpool.Pool
	escrowRepo  domainescrow.Repository
	postJournal *ledgeruc.PostJournal
}

// NewFundEscrow creates the FundEscrow use case.
func NewFundEscrow(pool *pgxpool.Pool, escrowRepo domainescrow.Repository, postJournal *ledgeruc.PostJournal) *FundEscrow {
	return &FundEscrow{pool: pool, escrowRepo: escrowRepo, postJournal: postJournal}
}

// Execute posts PSP_FUND_ESCROW and transitions escrow to FUNDED.
func (uc *FundEscrow) Execute(ctx context.Context, input FundEscrowInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" || input.PaymentOrderID == "" {
		return nil, domainescrow.ErrInvalidAmount
	}

	var result *domainescrow.Escrow
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		escrow, err := uc.escrowRepo.GetByShipmentID(txCtx, input.ShipmentID)
		if err != nil {
			return err
		}
		if !escrow.Status.CanFund() {
			return domainescrow.ErrInvalidTransition
		}

		_, err = uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypePSPFundEscrow,
			RefID:       fmt.Sprintf("%s:psp-fund:%s", input.ShipmentID, input.PaymentOrderID),
			Description: "PSP fund escrow",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.AccountIRPSPClearing, Debit: escrow.Amount, Credit: 0},
				{AccountCode: domainledger.ShipmentEscrowAccount(input.ShipmentID), Debit: 0, Credit: escrow.Amount},
			},
		})
		if err != nil {
			return err
		}

		now := nowUTC()
		escrow.Status = domainescrow.StatusFunded
		escrow.PaymentOrderID = input.PaymentOrderID
		escrow.FundingSource = domainescrow.FundingSourcePSP
		escrow.FundedAt = &now
		if err := uc.escrowRepo.Update(txCtx, escrow); err != nil {
			return err
		}
		result = escrow
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
