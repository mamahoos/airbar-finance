package idempotency

import (
	"errors"

	domainidempotency "github.com/mamahoos/airbar-finance/internal/domain/idempotency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validationStatus(err error) error {
	return status.Error(codes.InvalidArgument, err.Error())
}

func conflictStatus(err error) error {
	return status.Error(codes.Aborted, err.Error())
}

func notFoundStatus(err error) error {
	return status.Error(codes.NotFound, err.Error())
}

func MapDomainError(err error) error {
	switch {
	case domainidempotency.IsValidation(err):
		return validationStatus(err)
	case domainidempotency.IsConflict(err):
		return conflictStatus(err)
	case domainidempotency.IsNotFound(err):
		return notFoundStatus(err)
	default:
		return err
	}
}

func IsValidation(err error) bool {
	return domainidempotency.IsValidation(err)
}

func IsConflict(err error) bool {
	return domainidempotency.IsConflict(err)
}

func IsNotFound(err error) bool {
	return domainidempotency.IsNotFound(err)
}

func IsGRPCValidation(err error) bool {
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.InvalidArgument && errors.Is(err, domainidempotency.ErrKeyRequired)
}
