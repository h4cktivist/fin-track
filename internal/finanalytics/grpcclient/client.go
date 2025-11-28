package grpcclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"fin-track-app/api/proto"
	"fin-track-app/internal/domain"
)

type Client struct {
	client proto.TransactionServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("dial grpc server: %w", err)
	}
	return &Client{client: proto.NewTransactionServiceClient(conn)}, nil
}

func (c *Client) FetchTransactions(ctx context.Context, userID string) ([]domain.Transaction, error) {
	response, err := c.client.GetUserTransactions(ctx, &proto.UserRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("grpc get transactions: %w", err)
	}

	transactions := make([]domain.Transaction, 0, len(response.GetTransactions()))
	for _, tx := range response.GetTransactions() {
		parsed, _ := time.Parse(time.RFC3339, tx.GetCreatedAt())
		transactions = append(transactions, domain.Transaction{
			ID:        tx.GetId(),
			UserID:    tx.GetUserId(),
			Amount:    tx.GetAmount(),
			Category:  tx.GetCategory(),
			Type:      domain.TransactionType(tx.GetType()),
			CreatedAt: parsed,
		})
	}

	return transactions, nil
}
