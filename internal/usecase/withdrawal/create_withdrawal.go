package withdrawal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
	walletuc "github.com/mamahoos/airbar-finance/internal/usecase/wallet"
	audituc "github.com/mamahoos/airbar-finance/internal/usecase/audit"
)

// CreateWithdrawalInput is the application input for UC-16.
type CreateWithdrawalInput struct {
	UserID               string
	Amount               int64
	DestinationIBAN      string
	UserActive           bool
	FinancialKycApproved bool
}

// CreateWithdrawal reserves wallet funds for payout.
type CreateWithdrawal struct {
	pool        *pgxpool.Pool
	withdrawals domainwithdrawal.Repository
	postJournal *ledgeruc.PostJournal
	getBalance  *walletuc.GetBalance
	audit       *audituc.Emitter
}

// NewCreateWithdrawal creates the CreateWithdrawal use case.
func NewCreateWithdrawal(
	pool *pgxpool.Pool,
	withdrawals domainwithdrawal.Repository,
	postJournal *ledgeruc.PostJournal,
	getBalance *walletuc.GetBalance,
	audit *audituc.Emitter,
) *CreateWithdrawal {
	return &CreateWithdrawal{
		pool:        pool,
		withdrawals: withdrawals,
		postJournal: postJournal,
		getBalance:  getBalance,
		audit:       audit,
	}
}

// Execute validates gates, posts WITHDRAWAL_RESERVE, and creates a PENDING withdrawal.
func (uc *CreateWithdrawal) Execute(ctx context.Context, input CreateWithdrawalInput) (*domainwithdrawal.Withdrawal, error) {
	if input.UserID == "" || input.Amount <= 0 || input.DestinationIBAN == "" {
		return nil, domainwithdrawal.ErrInvalidInput
	}
	if !input.UserActive {
		return nil, domainwithdrawal.ErrUserInactive
	}
	if !input.FinancialKycApproved {
		return nil, domainwithdrawal.ErrKycNotApproved
	}

	balance, err := uc.getBalance.Execute(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if balance < input.Amount {
		return nil, domainwithdrawal.ErrInsufficientWallet
	}

	withdrawal := &domainwithdrawal.Withdrawal{
		UserID:          input.UserID,
		Amount:          input.Amount,
		Status:          domainwithdrawal.StatusPending,
		DestinationHash: domainwithdrawal.HashDestination(input.DestinationIBAN),
	}

	var result *domainwithdrawal.Withdrawal
	err = pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		if err := uc.withdrawals.Create(txCtx, withdrawal); err != nil {
			return err
		}
		_ = uc.audit.EmitWithdrawalCreated(txCtx, withdrawal.ID, withdrawal.UserID, string(withdrawal.Status))

		_, err := uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypeWithdrawalReserve,
			RefID:       fmt.Sprintf("%s:reserve", withdrawal.ID),
			Description: "Withdrawal reserve",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.UserWalletAccount(input.UserID), Debit: input.Amount, Credit: 0},
				{AccountCode: domainledger.AccountIRPayoutClearing, Debit: 0, Credit: input.Amount},
			},
		})
		if err != nil {
			return err
		}

		result = withdrawal
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

var nowUTC = func() time.Time { return time.Now().UTC() }
