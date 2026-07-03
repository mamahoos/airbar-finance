package escrow

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
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
	audit       *audituc.Emitter
}

// NewFundEscrow creates the FundEscrow use case.
func NewFundEscrow(pool *pgxpool.Pool, escrowRepo domainescrow.Repository, postJournal *ledgeruc.PostJournal, audit *audituc.Emitter) *FundEscrow {
	return &FundEscrow{pool: pool, escrowRepo: escrowRepo, postJournal: postJournal, audit: audit}
}

// Execute routes PSP payment through the payer wallet and transitions escrow to FUNDED.
func (uc *FundEscrow) Execute(ctx context.Context, input FundEscrowInput) (*domainescrow.Escrow, error) {
	if input.ShipmentID == "" || input.PaymentOrderID == "" {
		return nil, domainescrow.ErrInvalidAmount
	}

	if _, ok := pg.TxFromContext(ctx); ok {
		return uc.execute(ctx, input)
	}

	var result *domainescrow.Escrow
	err := pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		escrow, err := uc.execute(txCtx, input)
		if err != nil {
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

func (uc *FundEscrow) execute(ctx context.Context, input FundEscrowInput) (*domainescrow.Escrow, error) {
	escrow, err := uc.escrowRepo.GetByShipmentID(ctx, input.ShipmentID)
	if err != nil {
		return nil, err
	}
	if !escrow.Status.CanFund() {
		return nil, domainescrow.ErrInvalidTransition
	}

	// Direct payment is routed through the payer wallet in two legs so that
	// the wallet transaction history reflects both the incoming PSP credit and
	// the outgoing escrow funding. The net wallet balance change is zero, so no
	// withdrawable cash is created.
	_, err = uc.postJournal.Execute(ctx, ledgeruc.PostJournalInput{
		RefType:     domainledger.RefTypePSPToWallet,
		RefID:       fmt.Sprintf("%s:psp-wallet:%s", input.ShipmentID, input.PaymentOrderID),
		Description: "PSP credit to payer wallet",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.AccountIRPSPClearing, Debit: escrow.Amount, Credit: 0},
			{AccountCode: domainledger.UserWalletAccount(escrow.PayerUserID), Debit: 0, Credit: escrow.Amount},
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = uc.postJournal.Execute(ctx, ledgeruc.PostJournalInput{
		RefType:     domainledger.RefTypeWalletToEscrow,
		RefID:       fmt.Sprintf("%s:wallet-escrow:%s", input.ShipmentID, input.PaymentOrderID),
		Description: "Fund escrow from payer wallet",
		Lines: []domainledger.EntryLine{
			{AccountCode: domainledger.UserWalletAccount(escrow.PayerUserID), Debit: escrow.Amount, Credit: 0},
			{AccountCode: domainledger.ShipmentEscrowAccount(input.ShipmentID), Debit: 0, Credit: escrow.Amount},
		},
	})
	if err != nil {
		return nil, err
	}

	now := nowUTC()
	escrow.Status = domainescrow.StatusFunded
	escrow.PaymentOrderID = input.PaymentOrderID
	escrow.FundingSource = domainescrow.FundingSourcePSP
	escrow.FundedAt = &now
	if err := uc.escrowRepo.Update(ctx, escrow); err != nil {
		return nil, err
	}
	emitEscrowStatus(ctx, uc.audit, escrow)
	return escrow, nil
}
