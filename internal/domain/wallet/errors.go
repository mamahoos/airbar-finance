package wallet

import "errors"

var (
	// ErrInvalidInput is returned when required fields are missing or invalid.
	ErrInvalidInput = errors.New("wallet: invalid input")
	// ErrUnsupportedCurrency is returned when currency is not IRT.
	ErrUnsupportedCurrency = errors.New("wallet: unsupported currency")
)
