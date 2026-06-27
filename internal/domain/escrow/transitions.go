package escrow

// CanPayFromWallet reports whether wallet funding is allowed.
func (s Status) CanPayFromWallet() bool {
	return s == StatusCreated
}

// CanFund reports whether PSP funding is allowed.
func (s Status) CanFund() bool {
	return s == StatusCreated
}

// CanMarkDelivered reports whether delivery can start the dispute window.
func (s Status) CanMarkDelivered() bool {
	return s == StatusFunded
}

// CanFreeze reports whether the escrow can be frozen for dispute.
func (s Status) CanFreeze() bool {
	return s == StatusDisputeWindow || s == StatusLocked
}

// CanRelease reports whether funds can be released to the carrier.
func (s Status) CanRelease() bool {
	return s == StatusDisputeWindow || s == StatusPartiallyRefunded
}

// CanRefund reports whether a full refund to the payer wallet is allowed.
func (s Status) CanRefund() bool {
	switch s {
	case StatusFunded, StatusDisputeWindow, StatusFrozen, StatusPartiallyRefunded:
		return true
	default:
		return false
	}
}

// CanPartialRefund reports whether a partial refund is allowed.
func (s Status) CanPartialRefund() bool {
	return s.CanRefund()
}
