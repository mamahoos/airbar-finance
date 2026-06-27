package escrow

import (
	"strconv"
	"time"

	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
)

// ParseAmount parses a decimal string amount in rials.
func ParseAmount(raw string) (int64, error) {
	if raw == "" {
		return 0, domainescrow.ErrInvalidAmount
	}
	amount, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || amount <= 0 {
		return 0, domainescrow.ErrInvalidAmount
	}
	return amount, nil
}

// FormatAmount formats rials as a decimal string for proto responses.
func FormatAmount(amount int64) string {
	return strconv.FormatInt(amount, 10)
}

// FormatTimestamp converts an optional timestamp for proto mapping.
func FormatTimestamp(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	t := value.UTC()
	return &t
}
