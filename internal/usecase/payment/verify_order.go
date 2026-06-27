package payment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	domainledger "github.com/mamahoos/airbar-finance/internal/domain/ledger"
	domainpayment "github.com/mamahoos/airbar-finance/internal/domain/payment"
	domainprovider "github.com/mamahoos/airbar-finance/internal/domain/provider"
	pg "github.com/mamahoos/airbar-finance/internal/infrastructure/postgres"
	escrowuc "github.com/mamahoos/airbar-finance/internal/usecase/escrow"
	ledgeruc "github.com/mamahoos/airbar-finance/internal/usecase/ledger"
)

// VerifyOrderInput is input for confirming a payment order with Zibal.
type VerifyOrderInput struct {
	OrderID   string
	Authority string
	Purpose   domainpayment.Purpose
}

// VerifyOrder verifies a Zibal payment and completes ledger side effects.
type VerifyOrder struct {
	pool        *pgxpool.Pool
	orders      domainpayment.Repository
	events      domainprovider.Repository
	zibal       ZibalGateway
	fundEscrow  *escrowuc.FundEscrow
	postJournal *ledgeruc.PostJournal
}

// NewVerifyOrder creates the VerifyOrder use case.
func NewVerifyOrder(
	pool *pgxpool.Pool,
	orders domainpayment.Repository,
	events domainprovider.Repository,
	zibalClient ZibalGateway,
	fundEscrow *escrowuc.FundEscrow,
	postJournal *ledgeruc.PostJournal,
) *VerifyOrder {
	return &VerifyOrder{
		pool:        pool,
		orders:      orders,
		events:      events,
		zibal:       zibalClient,
		fundEscrow:  fundEscrow,
		postJournal: postJournal,
	}
}

// Execute verifies with Zibal and funds escrow or credits wallet.
func (uc *VerifyOrder) Execute(ctx context.Context, input VerifyOrderInput) (*domainpayment.Order, error) {
	order, err := uc.loadOrder(ctx, input)
	if err != nil {
		return nil, err
	}
	if input.Purpose != "" && order.Purpose != input.Purpose {
		return nil, domainpayment.ErrInvalidPurpose
	}
	if order.Status == domainpayment.StatusConfirmed {
		return order, nil
	}
	if order.Status == domainpayment.StatusFailed {
		return nil, domainpayment.ErrProviderVerifyFailed
	}

	verify, err := uc.zibal.Verify(ctx, order.Authority)
	if err != nil {
		return nil, domainpayment.ErrProviderVerifyFailed
	}
	if verify.Amount != order.Amount {
		return nil, domainpayment.ErrAmountMismatch
	}

	var result *domainpayment.Order
	err = pg.WithTx(ctx, uc.pool, func(txCtx context.Context) error {
		current, err := uc.orders.GetByID(txCtx, order.ID)
		if err != nil {
			return err
		}
		if current.Status == domainpayment.StatusConfirmed {
			result = current
			return nil
		}

		switch current.Purpose {
		case domainpayment.PurposeShipment:
			if _, err := uc.fundEscrow.Execute(txCtx, escrowuc.FundEscrowInput{
				ShipmentID:     current.ShipmentID,
				PaymentOrderID: current.ID,
			}); err != nil {
				return err
			}
		case domainpayment.PurposeWalletTopup:
			_, err := uc.postJournal.Execute(txCtx, ledgeruc.PostJournalInput{
				RefType:     domainledger.RefTypeWalletTopup,
				RefID:       fmt.Sprintf("%s:topup", current.ID),
				Description: "Wallet topup via Zibal",
				Lines: []domainledger.EntryLine{
					{AccountCode: domainledger.AccountIRPSPClearing, Debit: current.Amount, Credit: 0},
					{AccountCode: domainledger.UserWalletAccount(current.PayerUserID), Debit: 0, Credit: current.Amount},
				},
			})
			if err != nil {
				return err
			}
		default:
			return domainpayment.ErrInvalidPurpose
		}

		now := time.Now().UTC()
		current.Status = domainpayment.StatusConfirmed
		current.VerifiedAt = &now
		if err := uc.orders.Update(txCtx, current); err != nil {
			return err
		}

		payload, hash, err := HashPayload(map[string]any{
			"trackId": current.Authority,
			"orderId": current.ID,
			"amount":  verify.Amount,
		})
		if err != nil {
			return err
		}
		if err := uc.events.Create(txCtx, &domainprovider.Event{
			Provider:       domainprovider.ProviderZibal,
			EventType:      domainprovider.EventTypeVerify,
			PaymentOrderID: current.ID,
			Payload:        payload,
			PayloadHash:    hash,
			IdempotencyKey: fmt.Sprintf("zibal:verify:%s", current.ID),
			Processed:      true,
		}); err != nil {
			return err
		}

		result = current
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (uc *VerifyOrder) loadOrder(ctx context.Context, input VerifyOrderInput) (*domainpayment.Order, error) {
	if input.OrderID != "" {
		return uc.orders.GetByID(ctx, input.OrderID)
	}
	authority := strings.TrimSpace(input.Authority)
	if authority == "" {
		return nil, domainpayment.ErrInvalidInput
	}
	return uc.orders.GetByAuthority(ctx, authority)
}

// VerifyPaymentOrder wraps VerifyOrder for shipment orders (UC-11).
type VerifyPaymentOrder struct {
	verify *VerifyOrder
}

// NewVerifyPaymentOrder creates UC-11 handler dependency.
func NewVerifyPaymentOrder(verify *VerifyOrder) *VerifyPaymentOrder {
	return &VerifyPaymentOrder{verify: verify}
}

// ExecuteByAuthority verifies an order loaded by Zibal trackId.
func (uc *VerifyOrder) ExecuteByAuthority(ctx context.Context, authority string) (*domainpayment.Order, error) {
	order, err := uc.orders.GetByAuthority(ctx, authority)
	if err != nil {
		return nil, err
	}
	return uc.Execute(ctx, VerifyOrderInput{
		OrderID: order.ID,
		Purpose: order.Purpose,
	})
}

// Execute verifies a shipment payment order.
func (uc *VerifyPaymentOrder) Execute(ctx context.Context, orderID, authority string) (*domainpayment.Order, error) {
	return uc.verify.Execute(ctx, VerifyOrderInput{
		OrderID:   orderID,
		Authority: authority,
		Purpose:   domainpayment.PurposeShipment,
	})
}

// VerifyWalletTopupOrder wraps VerifyOrder for wallet topup (UC-13).
type VerifyWalletTopupOrder struct {
	verify *VerifyOrder
}

// NewVerifyWalletTopupOrder creates UC-13 handler dependency.
func NewVerifyWalletTopupOrder(verify *VerifyOrder) *VerifyWalletTopupOrder {
	return &VerifyWalletTopupOrder{verify: verify}
}

// Execute verifies a wallet topup order.
func (uc *VerifyWalletTopupOrder) Execute(ctx context.Context, orderID, authority string) (*domainpayment.Order, error) {
	return uc.verify.Execute(ctx, VerifyOrderInput{
		OrderID:   orderID,
		Authority: authority,
		Purpose:   domainpayment.PurposeWalletTopup,
	})
}

// FailPaymentOrder marks an order failed by authority (callback failure path).
type FailPaymentOrder struct {
	orders domainpayment.Repository
	events domainprovider.Repository
}

// NewFailPaymentOrder creates the FailPaymentOrder use case.
func NewFailPaymentOrder(orders domainpayment.Repository, events domainprovider.Repository) *FailPaymentOrder {
	return &FailPaymentOrder{orders: orders, events: events}
}

// Execute marks the order failed when Zibal callback reports failure.
func (uc *FailPaymentOrder) Execute(ctx context.Context, authority string, successParam string) (*domainpayment.Order, error) {
	if authority == "" {
		return nil, domainpayment.ErrInvalidInput
	}

	order, err := uc.orders.GetByAuthority(ctx, authority)
	if err != nil {
		return nil, err
	}
	if order.Status == domainpayment.StatusConfirmed {
		return order, nil
	}

	order.Status = domainpayment.StatusFailed
	if err := uc.orders.Update(ctx, order); err != nil {
		return nil, err
	}

	payload, hash, err := HashPayload(map[string]any{
		"trackId": authority,
		"success": successParam,
	})
	if err != nil {
		return nil, err
	}
	_ = uc.events.Create(ctx, &domainprovider.Event{
		Provider:       domainprovider.ProviderZibal,
		EventType:      domainprovider.EventTypeCallback,
		PaymentOrderID: order.ID,
		Payload:        payload,
		PayloadHash:    hash,
		IdempotencyKey: fmt.Sprintf("zibal:callback:%s:%s", authority, successParam),
		Processed:      true,
	})

	return order, nil
}

// IsAlreadyConfirmed reports whether err is a benign duplicate verify.
func IsAlreadyConfirmed(order *domainpayment.Order, err error) bool {
	return err == nil && order != nil && order.Status == domainpayment.StatusConfirmed
}

// IsProviderFailure reports provider-side verify failures.
func IsProviderFailure(err error) bool {
	return errors.Is(err, domainpayment.ErrProviderVerifyFailed)
}
