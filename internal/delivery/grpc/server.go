package grpc

import (
	"fmt"
	"net"

	"github.com/mamahoos/airbar-finance/internal/delivery/grpc/handlers"
	grpcidempotency "github.com/mamahoos/airbar-finance/internal/delivery/grpc/idempotency"
	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/health"
	idempotencyuc "github.com/mamahoos/airbar-finance/internal/usecase/idempotency"
	"google.golang.org/grpc"
)

// Server wraps the gRPC server and its listener.
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer registers finance gRPC services and binds to port.
func NewServer(
	port int,
	checker *health.Checker,
	idempotencyGuard *idempotencyuc.Guard,
	escrowHandler *handlers.EscrowHandler,
	paymentHandler *handlers.PaymentHandler,
	walletHandler *handlers.WalletHandler,
	withdrawalHandler *handlers.WithdrawalHandler,
	treasuryHandler *handlers.TreasuryHandler,
	reconciliationHandler *handlers.ReconciliationHandler,
	providerEventHandler *handlers.ProviderEventHandler,
) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcidempotency.UnaryInterceptor(idempotencyGuard)),
	)
	financev1.RegisterFinanceHealthServiceServer(grpcServer, handlers.NewHealthHandler(checker))
	if escrowHandler != nil {
		financev1.RegisterEscrowServiceServer(grpcServer, escrowHandler)
	}
	if paymentHandler != nil {
		financev1.RegisterPaymentOrderServiceServer(grpcServer, paymentHandler)
	}
	if walletHandler != nil {
		financev1.RegisterWalletServiceServer(grpcServer, walletHandler)
	}
	if withdrawalHandler != nil {
		financev1.RegisterWithdrawalServiceServer(grpcServer, withdrawalHandler)
	}
	if treasuryHandler != nil {
		financev1.RegisterTreasuryServiceServer(grpcServer, treasuryHandler)
	}
	if reconciliationHandler != nil {
		financev1.RegisterReconciliationServiceServer(grpcServer, reconciliationHandler)
	}
	if providerEventHandler != nil {
		financev1.RegisterProviderEventServiceServer(grpcServer, providerEventHandler)
	}

	return &Server{
		grpcServer: grpcServer,
		listener:   listener,
	}, nil
}

// Serve starts accepting gRPC connections.
func (s *Server) Serve() error {
	return s.grpcServer.Serve(s.listener)
}

// GracefulStop stops the server gracefully.
func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

// Addr returns the bound listener address.
func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}
