package escrow

// CalcPlatformFee returns the platform fee in rials for a gross amount.
func CalcPlatformFee(amount int64, percent float64) int64 {
	if amount <= 0 {
		return 0
	}
	fee := int64(float64(amount) * percent / 100.0)
	if fee < 0 {
		return 0
	}
	if fee > amount {
		return amount
	}
	return fee
}
