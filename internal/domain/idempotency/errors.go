package idempotency

import "errors"

var (
	// ErrKeyRequired is returned when a mutating RPC has no idempotency key.
	ErrKeyRequired = errors.New("idempotency key required")
	// ErrConflict is returned when the same key is already in progress.
	ErrConflict = errors.New("idempotency conflict")
	// ErrNotFound is returned when an idempotency record does not exist.
	ErrNotFound = errors.New("idempotency record not found")
)

// IsValidation reports validation errors for mutating commands.
func IsValidation(err error) bool {
	return errors.Is(err, ErrKeyRequired)
}

// IsConflict reports duplicate in-flight idempotency keys.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsNotFound reports missing idempotency records.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
