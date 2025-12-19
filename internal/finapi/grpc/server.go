package grpc

import (
	"fmt"
	"net"

	stdgrpc "google.golang.org/grpc"

	"fin-track-app/api/proto"
	"fin-track-app/internal/finapi/service"
)

type Server struct {
	proto.UnimplementedTransactionServiceServer
	service *service.TransactionService
	server  *stdgrpc.Server
}

func NewServer(service *service.TransactionService) *Server {
	return &Server{
		service: service,
		server:  stdgrpc.NewServer(),
	}
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	proto.RegisterTransactionServiceServer(s.server, s)
	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}
