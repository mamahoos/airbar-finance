package handlers

import (
	"errors"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
	domainwallet "github.com/mamahoos/airbar-finance/internal/domain/wallet"
	domainwithdrawal "github.com/mamahoos/airbar-finance/internal/domain/withdrawal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapEscrowError(err error) error {
	switch {
	case errors.Is(err, domainescrow.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domainescrow.ErrDuplicateShipment):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domainescrow.ErrInvalidAmount),
		errors.Is(err, domainescrow.ErrAmountMismatch),
		errors.Is(err, domainescrow.ErrPayerMismatch):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domainescrow.ErrInvalidTransition),
		errors.Is(err, domainescrow.ErrInsufficientWallet),
		errors.Is(err, domainescrow.ErrNoEscrowBalance),
		errors.Is(err, domainescrow.ErrRefundExceedsBalance):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domainledger.ErrDuplicateJournal):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domainledger.ErrUnbalancedJournal),
		errors.Is(err, domainledger.ErrEmptyJournal),
		errors.Is(err, domainledger.ErrInvalidEntry):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func mapPaymentError(err error) error {
	switch {
	case errors.Is(err, domainpayment.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domainpayment.ErrInvalidInput),
		errors.Is(err, domainpayment.ErrAmountMismatch),
		errors.Is(err, domainpayment.ErrInvalidPurpose):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domainpayment.ErrEscrowNotReady),
		errors.Is(err, domainescrow.ErrPayerMismatch),
		errors.Is(err, domainpayment.ErrProviderVerifyFailed):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domainledger.ErrDuplicateJournal):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func mapWalletError(err error) error {
	switch {
	case errors.Is(err, domainwallet.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domainwallet.ErrUnsupportedCurrency):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func mapWithdrawalError(err error) error {
	switch {
	case errors.Is(err, domainwithdrawal.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domainwithdrawal.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domainwithdrawal.ErrInsufficientWallet),
		errors.Is(err, domainwithdrawal.ErrKycNotApproved),
		errors.Is(err, domainwithdrawal.ErrUserInactive),
		errors.Is(err, domainwithdrawal.ErrInvalidTransition):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domainledger.ErrDuplicateJournal):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
