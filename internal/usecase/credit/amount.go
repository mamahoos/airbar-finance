package credit

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseAmount parses a rials amount string into int64.
func ParseAmount(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, fmt.Errorf("empty amount")
	}
	amount, err := strconv.ParseInt(value, 10, 64)
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("invalid amount")
	}
	return amount, nil
}

// FormatAmount formats rials as a decimal string for proto responses.
func FormatAmount(amount int64) string {
	return strconv.FormatInt(amount, 10)
}
