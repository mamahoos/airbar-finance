package payment

import "errors"

var (
	// ErrNotFound is returned when a payment order does not exist.
	ErrNotFound = errors.New("payment: order not found")
	// ErrInvalidInput is returned when required fields are missing or invalid.
	ErrInvalidInput = errors.New("payment: invalid input")
	// ErrInvalidPurpose is returned when an operation does not match order purpose.
	ErrInvalidPurpose = errors.New("payment: invalid purpose")
	// ErrAmountMismatch is returned when verified amount differs from order amount.
	ErrAmountMismatch = errors.New("payment: amount mismatch")
	// ErrAlreadyConfirmed is returned when verifying an already confirmed order.
	ErrAlreadyConfirmed = errors.New("payment: order already confirmed")
	// ErrEscrowNotReady is returned when escrow is not CREATED for shipment pay.
	ErrEscrowNotReady = errors.New("payment: escrow not ready for funding")
	// ErrProviderVerifyFailed is returned when Zibal verify fails.
	ErrProviderVerifyFailed = errors.New("payment: provider verify failed")
)
