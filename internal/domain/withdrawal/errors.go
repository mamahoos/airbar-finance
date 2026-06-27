package withdrawal

import "errors"

var (
	// ErrNotFound is returned when a withdrawal does not exist.
	ErrNotFound = errors.New("withdrawal: not found")
	// ErrInvalidInput is returned when required fields are missing or invalid.
	ErrInvalidInput = errors.New("withdrawal: invalid input")
	// ErrInvalidTransition is returned when a status change is not allowed.
	ErrInvalidTransition = errors.New("withdrawal: invalid status transition")
	// ErrInsufficientWallet is returned when wallet balance is too low.
	ErrInsufficientWallet = errors.New("withdrawal: insufficient wallet balance")
	// ErrKycNotApproved is returned when payout KYC gates fail.
	ErrKycNotApproved = errors.New("withdrawal: financial kyc not approved")
	// ErrUserInactive is returned when user is not active.
	ErrUserInactive = errors.New("withdrawal: user inactive")
)
