package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"fin-track-app/internal/domain"
	cachemocks "fin-track-app/internal/finanalytics/cache/mocks"
	grpcmocks "fin-track-app/internal/finanalytics/grpcclient/mocks"
	"fin-track-app/internal/finanalytics/service"
)

type ServiceTestSuite struct {
	suite.Suite
	mockCache  *cachemocks.StatsCache
	mockClient *grpcmocks.TransactionClient
	service    *service.Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockCache = cachemocks.NewStatsCache(s.T())
	s.mockClient = grpcmocks.NewTransactionClient(s.T())
	s.service = service.New(s.mockCache, s.mockClient)
}

func (s *ServiceTestSuite) TestProcessKafkaMessageSuccess() {
	ctx := context.Background()
	userID := 1
	txs := []domain.Transaction{
		{ID: 1, UserID: userID, Amount: 100, Type: domain.TransactionTypeIncome, Category: "Salary"},
		{ID: 2, UserID: userID, Amount: 50, Type: domain.TransactionTypeExpense, Category: "Food"},
	}
	payload := domain.TransactionMessage{
		UserID:       userID,
		Transactions: txs,
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := &sarama.ConsumerMessage{
		Value: payloadBytes,
	}

	s.mockCache.On("Set", ctx, mock.MatchedBy(func(stats domain.FinanceStats) bool {
		return stats.UserID == userID && stats.TotalIncome == 100 && stats.TotalExpense == 50 && stats.Balance == 50
	})).Return(nil)

	err := s.service.ProcessKafkaMessage(ctx, msg)
	s.NoError(err)
}

func (s *ServiceTestSuite) TestGetStatsCached() {
	ctx := context.Background()
	userID := 1
	cachedStats := &domain.FinanceStats{
		UserID:      userID,
		TotalIncome: 1000,
	}

	s.mockCache.On("Get", ctx, userID).Return(cachedStats, nil)

	stats, err := s.service.GetStats(ctx, userID)
	s.NoError(err)
	s.Equal(1000.0, stats.TotalIncome)
}

func (s *ServiceTestSuite) TestGetStatsNotCachedSuccess() {
	ctx := context.Background()
	userID := 1
	txs := []domain.Transaction{
		{ID: 1, UserID: userID, Amount: 200, Type: domain.TransactionTypeIncome},
	}

	s.mockCache.On("Get", ctx, userID).Return(nil, errors.New("not found"))
	s.mockClient.On("FetchTransactions", ctx, userID).Return(txs, nil)
	s.mockCache.On("Set", ctx, mock.MatchedBy(func(stats domain.FinanceStats) bool {
		return stats.TotalIncome == 200
	})).Return(nil)

	stats, err := s.service.GetStats(ctx, userID)
	s.NoError(err)
	s.Equal(200.0, stats.TotalIncome)
}

func (s *ServiceTestSuite) TestGetStatsFetchError() {
	ctx := context.Background()
	userID := 1

	s.mockCache.On("Get", ctx, userID).Return(nil, errors.New("not found"))
	s.mockClient.On("FetchTransactions", ctx, userID).Return(nil, errors.New("fetch error"))

	_, err := s.service.GetStats(ctx, userID)
	s.Error(err)
	s.Equal("fetch error", err.Error())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
