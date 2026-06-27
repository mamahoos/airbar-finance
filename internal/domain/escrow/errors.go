package escrow

import "errors"

var (
	// ErrNotFound is returned when no escrow exists for the shipment.
	ErrNotFound = errors.New("escrow: not found")
	// ErrDuplicateShipment is returned when an escrow already exists for the shipment.
	ErrDuplicateShipment = errors.New("escrow: shipment already has escrow")
	// ErrInvalidTransition is returned when a status change is not allowed.
	ErrInvalidTransition = errors.New("escrow: invalid status transition")
	// ErrInvalidAmount is returned when an amount is zero or negative.
	ErrInvalidAmount = errors.New("escrow: invalid amount")
	// ErrAmountMismatch is returned when a command amount does not match escrow amount.
	ErrAmountMismatch = errors.New("escrow: amount mismatch")
	// ErrPayerMismatch is returned when payer_user_id does not match escrow.
	ErrPayerMismatch = errors.New("escrow: payer mismatch")
	// ErrInsufficientWallet is returned when payer wallet balance is too low.
	ErrInsufficientWallet = errors.New("escrow: insufficient wallet balance")
	// ErrNoEscrowBalance is returned when ledger escrow balance is zero.
	ErrNoEscrowBalance = errors.New("escrow: no funds in escrow account")
	// ErrRefundExceedsBalance is returned when partial refund exceeds escrow balance.
	ErrRefundExceedsBalance = errors.New("escrow: refund exceeds escrow balance")
)
