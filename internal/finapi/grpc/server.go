package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	stdgrpc "google.golang.org/grpc"

	"fin-track-app/api/proto"
	"fin-track-app/internal/domain"
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

func (s *Server) GetUserTransactions(ctx context.Context, req *proto.UserRequest) (*proto.UserTransactions, error) {
	items, err := s.service.ListTransactions(ctx, int(req.GetUserId()))
	if err != nil {
		return nil, err
	}

	return &proto.UserTransactions{
		Transactions: convertDomainTransactions(items),
	}, nil
}

func convertDomainTransactions(items []domain.Transaction) []*proto.Transaction {
	result := make([]*proto.Transaction, 0, len(items))
	for _, tx := range items {
		result = append(result, &proto.Transaction{
			Id:        tx.ID,
			UserId:    int64(tx.UserID),
			Amount:    tx.Amount,
			Category:  tx.Category,
			Type:      string(tx.Type),
			CreatedAt: tx.CreatedAt.Format(time.RFC3339),
		})
	}
	return result
}
