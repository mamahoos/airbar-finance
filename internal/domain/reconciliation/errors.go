package reconciliation

import "errors"

var (
	// ErrNotFound is returned when a reconciliation run id does not exist.
	ErrNotFound = errors.New("reconciliation run not found")
)
