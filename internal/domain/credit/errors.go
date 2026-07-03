package credit

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid credit input")
	ErrNotFound          = errors.New("credit grant not found")
	ErrDuplicateGrant    = errors.New("duplicate credit grant")
	ErrAlreadyReversed   = errors.New("credit grant already reversed")
	ErrUnsupportedCurrency = errors.New("unsupported credit currency")
)
