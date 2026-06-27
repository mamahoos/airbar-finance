package withdrawal

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HashDestination returns sha256 hex of a normalized IBAN. Plain IBAN must not be stored.
func HashDestination(iban string) string {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(iban), " ", ""))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
