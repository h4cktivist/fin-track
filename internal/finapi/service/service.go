package service

import (
	"context"
	"fmt"

	"fin-track-app/internal/domain"
	repo "fin-track-app/internal/finapi/repository"
	publisher "fin-track-app/internal/kafka"
)

type TransactionService struct {
	repo      repo.TransactionRepository
	publisher publisher.EventPublisher
}

func NewTransactionService(repo repo.TransactionRepository, publisher publisher.EventPublisher) *TransactionService {
	return &TransactionService{
		repo:      repo,
		publisher: publisher,
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

func (s *TransactionService) ListTransactions(ctx context.Context, userID int) ([]domain.Transaction, error) {
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

func (s *TransactionService) DeleteTransaction(ctx context.Context, userID int, transactionID int64) error {
	if err := s.repo.DeleteTransaction(ctx, userID, transactionID); err != nil {
		return err
	}

	return s.publishUserTransactions(ctx, userID)
}

func (s *TransactionService) publishUserTransactions(ctx context.Context, userID int) error {
	all, err := s.repo.ListUserTransactions(ctx, userID)
	if err != nil {
		return fmt.Errorf("list user transactions: %w", err)
	}

	if err := s.publisher.PublishTransactions(ctx, domain.TransactionMessage{
		UserID:       userID,
		Transactions: all,
	}); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}
