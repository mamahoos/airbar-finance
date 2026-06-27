package payment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"

	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
)

// ParseAmount parses a decimal string amount in rials.
func ParseAmount(raw string) (int64, error) {
	if raw == "" {
		return 0, domainpayment.ErrInvalidInput
	}
	amount, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || amount <= 0 {
		return 0, domainpayment.ErrInvalidInput
	}
	return amount, nil
}

// FormatAmount formats rials as a decimal string for proto responses.
func FormatAmount(amount int64) string {
	return strconv.FormatInt(amount, 10)
}

// HashPayload returns sha256 hex of JSON payload bytes.
func HashPayload(payload any) ([]byte, string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(raw)
	return raw, hex.EncodeToString(sum[:]), nil
}

// CallbackURL builds the finance Zibal callback URL.
func CallbackURL(publicBaseURL string) string {
	return publicBaseURL + "/api/v1/zibal/callback"
}
