package service_test

import (
	"context"
	"errors"
	repomocks "fin-api/internal/repository/mocks"
	"fin-api/internal/service"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"fin-api/internal/domain"
	kafkamocks "fin-api/internal/kafka/mocks"
)

type TransactionServiceTestSuite struct {
	suite.Suite
	mockRepo      *repomocks.TransactionRepository
	mockPublisher *kafkamocks.EventPublisher
	service       *service.TransactionService
}

func (s *TransactionServiceTestSuite) SetupTest() {
	s.mockRepo = repomocks.NewTransactionRepository(s.T())
	s.mockPublisher = kafkamocks.NewEventPublisher(s.T())
	s.service = service.NewTransactionService(s.mockRepo, s.mockPublisher)
}

func (s *TransactionServiceTestSuite) TestCreateTransactionSuccess() {
	ctx := context.Background()
	userID := 1
	tx := domain.Transaction{
		UserID:   userID,
		Amount:   100,
		Category: "Salary",
		Type:     domain.TransactionTypeIncome,
	}
	createdTx := tx
	createdTx.ID = 1

	s.mockRepo.On("CreateTransaction", ctx, tx).Return(createdTx, nil)
	s.mockRepo.On("ListUserTransactions", ctx, userID).Return([]domain.Transaction{createdTx}, nil)
	s.mockPublisher.On("PublishTransactions", ctx, mock.MatchedBy(func(msg domain.TransactionMessage) bool {
		return msg.UserID == userID && len(msg.Transactions) == 1 && msg.Transactions[0].ID == 1
	})).Return(nil)

	result, err := s.service.CreateTransaction(ctx, tx)
	s.NoError(err)
	s.Equal(createdTx, result)
}

func (s *TransactionServiceTestSuite) TestCreateTransactionRepoError() {
	ctx := context.Background()
	tx := domain.Transaction{UserID: 1}

	s.mockRepo.On("CreateTransaction", ctx, tx).Return(domain.Transaction{}, errors.New("repo error"))

	_, err := s.service.CreateTransaction(ctx, tx)
	s.Error(err)
	s.Equal("repo error", err.Error())
}

func (s *TransactionServiceTestSuite) TestListTransactionsSuccess() {
	ctx := context.Background()
	userID := 1
	txs := []domain.Transaction{{ID: 1, UserID: userID}}

	s.mockRepo.On("ListUserTransactions", ctx, userID).Return(txs, nil)

	result, err := s.service.ListTransactions(ctx, userID)
	s.NoError(err)
	s.Equal(txs, result)
}

func (s *TransactionServiceTestSuite) TestUpdateTransactionSuccess() {
	ctx := context.Background()
	userID := 1
	tx := domain.Transaction{ID: 1, UserID: userID, Amount: 150}
	updatedTx := tx

	s.mockRepo.On("UpdateTransaction", ctx, tx).Return(updatedTx, nil)
	s.mockRepo.On("ListUserTransactions", ctx, userID).Return([]domain.Transaction{updatedTx}, nil)
	s.mockPublisher.On("PublishTransactions", ctx, mock.Anything).Return(nil)

	result, err := s.service.UpdateTransaction(ctx, tx)
	s.NoError(err)
	s.Equal(updatedTx, result)
}

func (s *TransactionServiceTestSuite) TestUpdateTransactionRepoError() {
	ctx := context.Background()
	userID := 1
	tx := domain.Transaction{ID: 999, UserID: userID, Amount: 150}

	s.mockRepo.On("UpdateTransaction", ctx, tx).Return(domain.Transaction{}, errors.New("transaction not found"))

	_, err := s.service.UpdateTransaction(ctx, tx)
	s.Error(err)
	s.Equal("transaction not found", err.Error())
}

func (s *TransactionServiceTestSuite) TestDeleteTransactionSuccess() {
	ctx := context.Background()
	userID := 1
	txID := int64(1)

	s.mockRepo.On("DeleteTransaction", ctx, userID, txID).Return(nil)
	s.mockRepo.On("ListUserTransactions", ctx, userID).Return([]domain.Transaction{}, nil)
	s.mockPublisher.On("PublishTransactions", ctx, mock.Anything).Return(nil)

	err := s.service.DeleteTransaction(ctx, userID, txID)
	s.NoError(err)
}

func (s *TransactionServiceTestSuite) TestDeleteTransactionRepoError() {
	ctx := context.Background()
	userID := 1
	txID := int64(999)

	s.mockRepo.On("DeleteTransaction", ctx, userID, txID).Return(errors.New("transaction not found"))

	err := s.service.DeleteTransaction(ctx, userID, txID)
	s.Error(err)
	s.Equal("transaction not found", err.Error())
}

func TestTransactionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionServiceTestSuite))
}
