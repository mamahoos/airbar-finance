package grpc

import (
	"fmt"
	"net"

	"github.com/mamahoos/airbar-finance/internal/delivery/grpc/handlers"
	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	"github.com/mamahoos/airbar-finance/internal/infrastructure/health"
	"google.golang.org/grpc"
)

// Server wraps the gRPC server and its listener.
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer registers finance gRPC services and binds to port.
func NewServer(port int, checker *health.Checker, escrowHandler *handlers.EscrowHandler, paymentHandler *handlers.PaymentHandler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	financev1.RegisterFinanceHealthServiceServer(grpcServer, handlers.NewHealthHandler(checker))
	if escrowHandler != nil {
		financev1.RegisterEscrowServiceServer(grpcServer, escrowHandler)
	}
	if paymentHandler != nil {
		financev1.RegisterPaymentOrderServiceServer(grpcServer, paymentHandler)
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
