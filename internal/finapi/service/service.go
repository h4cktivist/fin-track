package service

import (
	"context"
	"fmt"

	"fin-track-app/internal/domain"
	"fin-track-app/internal/finapi/repository"
	"fin-track-app/internal/kafka"
)

type TransactionService struct {
	repo     *repository.TransactionRepository
	producer *kafka.Producer
}

func NewTransactionService(repo *repository.TransactionRepository, producer *kafka.Producer) *TransactionService {
	return &TransactionService{
		repo:     repo,
		producer: producer,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	created, err := s.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return domain.Transaction{}, err
	}

	if err := s.publishUserTransactions(ctx, tx.UserID); err != nil {
		return domain.Transaction{}, err
	}

	return created, nil
}

func (s *TransactionService) ListTransactions(ctx context.Context, userID string) ([]domain.Transaction, error) {
	return s.repo.ListUserTransactions(ctx, userID)
}

func (s *TransactionService) UpdateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	updated, err := s.repo.UpdateTransaction(ctx, tx)
	if err != nil {
		return domain.Transaction{}, err
	}

	if err := s.publishUserTransactions(ctx, tx.UserID); err != nil {
		return domain.Transaction{}, err
	}

	return updated, nil
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, userID string, transactionID int64) error {
	if err := s.repo.DeleteTransaction(ctx, userID, transactionID); err != nil {
		return err
	}

	return s.publishUserTransactions(ctx, userID)
}

func (s *TransactionService) publishUserTransactions(ctx context.Context, userID string) error {
	all, err := s.repo.ListUserTransactions(ctx, userID)
	if err != nil {
		return fmt.Errorf("list user transactions: %w", err)
	}

	if err := s.producer.PublishTransactions(ctx, domain.TransactionMessage{
		UserID:       userID,
		Transactions: all,
	}); err != nil {
		return fmt.Errorf("publish kafka message: %w", err)
	}

	return nil
}
