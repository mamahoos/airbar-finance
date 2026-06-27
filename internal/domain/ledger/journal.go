package ledger

import "time"

// RefType classifies the business event that produced a journal.
type RefType string

const (
	RefTypePSPFundEscrow            RefType = "PSP_FUND_ESCROW"
	RefTypeWalletToEscrow           RefType = "WALLET_TO_ESCROW"
	RefTypeEscrowRelease            RefType = "ESCROW_RELEASE"
	RefTypeEscrowRefundWallet       RefType = "ESCROW_REFUND_WALLET"
	RefTypeWalletTopup              RefType = "WALLET_TOPUP"
	RefTypeWithdrawalReserve        RefType = "WITHDRAWAL_RESERVE"
	RefTypeWithdrawalRejectReversal RefType = "WITHDRAWAL_REJECT_REVERSAL"
)

// EntryLine is input for constructing a journal entry before persistence.
type EntryLine struct {
	AccountCode AccountCode
	Debit       int64
	Credit      int64
}

// Entry is a persisted ledger line belonging to a journal.
type Entry struct {
	ID          string
	JournalID   string
	AccountCode AccountCode
	Debit       int64
	Credit      int64
	CreatedAt   time.Time
}

// Journal groups balanced entry lines for one business event.
type Journal struct {
	ID          string
	RefType     RefType
	RefID       string
	Description string
	Entries     []Entry
	CreatedAt   time.Time
}

// ValidateLines checks double-entry invariants for entry lines.
func ValidateLines(lines []EntryLine) error {
	if len(lines) == 0 {
		return ErrEmptyJournal
	}

	var totalDebit int64
	var totalCredit int64

	for _, line := range lines {
		if line.Debit < 0 || line.Credit < 0 {
			return ErrInvalidEntry
		}
		if (line.Debit > 0 && line.Credit > 0) || (line.Debit == 0 && line.Credit == 0) {
			return ErrInvalidEntry
		}
		if line.AccountCode == "" {
			return ErrInvalidEntry
		}
		totalDebit += line.Debit
		totalCredit += line.Credit
	}

	if totalDebit != totalCredit {
		return ErrUnbalancedJournal
	}

	return nil
}

// NewJournal builds an in-memory journal after validating entry lines.
func NewJournal(refType RefType, refID, description string, lines []EntryLine) (*Journal, error) {
	if refType == "" || refID == "" {
		return nil, ErrInvalidEntry
	}
	if err := ValidateLines(lines); err != nil {
		return nil, err
	}

	entries := make([]Entry, len(lines))
	for i, line := range lines {
		entries[i] = Entry{
			AccountCode: line.AccountCode,
			Debit:       line.Debit,
			Credit:      line.Credit,
		}
	}

	return &Journal{
		RefType:     refType,
		RefID:       refID,
		Description: description,
		Entries:     entries,
	}, nil
}
