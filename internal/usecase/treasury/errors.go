package treasury

import "errors"

// ErrUnsupportedCurrency is returned when treasury summary is requested for a non-IRT currency.
var ErrUnsupportedCurrency = errors.New("unsupported currency")
