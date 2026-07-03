package escrow

import (
	domainescrow "github.com/mamahoos/airbar-finance/internal/domain/escrow"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
)

// PayerFundingSplit is how an escrow payment is sourced from promo credit and wallet.
type PayerFundingSplit struct {
	PromoCredit int64
	Wallet      int64
}

// Total returns the combined payer-funded amount.
func (s PayerFundingSplit) Total() int64 {
	return s.PromoCredit + s.Wallet
}

// ResolvePayerFundingSplit computes promo-first split spend for an escrow payment.
func ResolvePayerFundingSplit(amount, promoBalance, walletBalance int64) (PayerFundingSplit, error) {
	if amount <= 0 {
		return PayerFundingSplit{}, domainescrow.ErrInvalidAmount
	}

	promoUsed := promoBalance
	if promoUsed > amount {
		promoUsed = amount
	}
	walletNeeded := amount - promoUsed
	if walletBalance < walletNeeded {
		return PayerFundingSplit{}, domainescrow.ErrInsufficientPayerFunds
	}

	return PayerFundingSplit{PromoCredit: promoUsed, Wallet: walletNeeded}, nil
}

// FundingSourceForSplit maps a payer split to the escrow funding_source label.
func FundingSourceForSplit(split PayerFundingSplit) domainescrow.FundingSource {
	switch {
	case split.PromoCredit > 0 && split.Wallet > 0:
		return domainescrow.FundingSourceMixed
	case split.PromoCredit > 0:
		return domainescrow.FundingSourcePromoCredit
	default:
		return domainescrow.FundingSourceWallet
	}
}

// BuildEscrowPayJournalLines builds balanced ledger lines for promo-first escrow funding.
func BuildEscrowPayJournalLines(payerUserID, shipmentID string, split PayerFundingSplit) []domainledger.EntryLine {
	lines := make([]domainledger.EntryLine, 0, 3)
	if split.PromoCredit > 0 {
		lines = append(lines,
			domainledger.EntryLine{AccountCode: domainledger.UserPromoCreditAccount(payerUserID), Debit: split.PromoCredit, Credit: 0},
		)
	}
	if split.Wallet > 0 {
		lines = append(lines,
			domainledger.EntryLine{AccountCode: domainledger.UserWalletAccount(payerUserID), Debit: split.Wallet, Credit: 0},
		)
	}
	lines = append(lines,
		domainledger.EntryLine{AccountCode: domainledger.ShipmentEscrowAccount(shipmentID), Debit: 0, Credit: split.Total()},
	)
	return lines
}

// RefTypeForEscrowPay picks the journal ref type for a payer-funded escrow payment.
func RefTypeForEscrowPay(split PayerFundingSplit) domainledger.RefType {
	if split.PromoCredit > 0 && split.Wallet > 0 {
		return domainledger.RefTypePromoCreditToEscrow
	}
	if split.PromoCredit > 0 {
		return domainledger.RefTypePromoCreditToEscrow
	}
	return domainledger.RefTypeWalletToEscrow
}

// RefundAllocation splits a refund between promo credit and wallet based on remaining promo-funded portion.
func RefundAllocation(refundAmount, promoCreditRemaining int64) (promoRefund, walletRefund int64) {
	if refundAmount <= 0 || promoCreditRemaining <= 0 {
		return 0, refundAmount
	}
	promoRefund = refundAmount
	if promoRefund > promoCreditRemaining {
		promoRefund = promoCreditRemaining
	}
	return promoRefund, refundAmount - promoRefund
}

// BuildEscrowRefundJournalLines builds refund journals returning funds to promo credit first, then wallet.
func BuildEscrowRefundJournalLines(
	payerUserID, shipmentID string,
	promoRefund, walletRefund int64,
) []ledgeruc.PostJournalInput {
	inputs := make([]ledgeruc.PostJournalInput, 0, 2)
	if promoRefund > 0 {
		inputs = append(inputs, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypeEscrowRefundPromoCredit,
			RefID:       shipmentID + ":refund-promo:" + FormatAmount(promoRefund),
			Description: "Refund escrow promo credit to payer",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.ShipmentEscrowAccount(shipmentID), Debit: promoRefund, Credit: 0},
				{AccountCode: domainledger.UserPromoCreditAccount(payerUserID), Debit: 0, Credit: promoRefund},
			},
		})
	}
	if walletRefund > 0 {
		inputs = append(inputs, ledgeruc.PostJournalInput{
			RefType:     domainledger.RefTypeEscrowRefundWallet,
			RefID:       shipmentID + ":refund-wallet:" + FormatAmount(walletRefund),
			Description: "Refund escrow to payer wallet",
			Lines: []domainledger.EntryLine{
				{AccountCode: domainledger.ShipmentEscrowAccount(shipmentID), Debit: walletRefund, Credit: 0},
				{AccountCode: domainledger.UserWalletAccount(payerUserID), Debit: 0, Credit: walletRefund},
			},
		})
	}
	return inputs
}
