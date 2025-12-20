package grpc

import (
	"context"
	"time"

	"fin-api/api/proto"
	"fin-api/internal/domain"
)

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
