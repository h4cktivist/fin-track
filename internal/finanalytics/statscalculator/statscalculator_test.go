package statscalculator

import (
	"testing"
	"time"

	"fin-track-app/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CalculatorTestSuite struct {
	suite.Suite
	transactions []domain.Transaction
}

func (s *CalculatorTestSuite) SetupTest() {
	s.transactions = []domain.Transaction{
		{
			ID:        1,
			UserID:    "user-123",
			Amount:    1000.0,
			Type:      domain.TransactionTypeIncome,
			Category:  "salary",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        2,
			UserID:    "user-123",
			Amount:    500.0,
			Type:      domain.TransactionTypeIncome,
			Category:  "bonus",
			CreatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID:        3,
			UserID:    "user-123",
			Amount:    300.0,
			Type:      domain.TransactionTypeExpense,
			Category:  "food",
			CreatedAt: time.Now().Add(-6 * time.Hour),
		},
		{
			ID:        4,
			UserID:    "user-123",
			Amount:    200.0,
			Type:      domain.TransactionTypeExpense,
			Category:  "transport",
			CreatedAt: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:        5,
			UserID:    "user-123",
			Amount:    150.0,
			Type:      domain.TransactionTypeExpense,
			Category:  "food",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}
}

func (s *CalculatorTestSuite) TestCalculateStats_EmptyTransactions() {
	var emptyTransactions []domain.Transaction

	stats := CalculateStats(emptyTransactions)

	assert.Equal(s.T(), "", stats.UserID)
	assert.Equal(s.T(), 0, stats.TransactionsCount)
	assert.Equal(s.T(), 0.0, stats.TotalIncome)
	assert.Equal(s.T(), 0.0, stats.TotalExpense)
	assert.Equal(s.T(), 0.0, stats.Balance)
	assert.Equal(s.T(), 0.0, stats.AverageIncome)
	assert.Equal(s.T(), 0.0, stats.AverageExpense)
	assert.Empty(s.T(), stats.IncomeByCategory)
	assert.Empty(s.T(), stats.ExpenseByCategory)
	assert.WithinDuration(s.T(), time.Now().UTC(), stats.GeneratedAt, time.Second)
}

func (s *CalculatorTestSuite) TestCalculateStats_MixedTransactions() {
	stats := CalculateStats(s.transactions)

	assert.Equal(s.T(), "user-123", stats.UserID)
	assert.Equal(s.T(), 5, stats.TransactionsCount)
	assert.Equal(s.T(), 1500.0, stats.TotalIncome)
	assert.Equal(s.T(), 650.0, stats.TotalExpense)
	assert.Equal(s.T(), 850.0, stats.Balance)
	assert.Equal(s.T(), 750.0, stats.AverageIncome)
	assert.Equal(s.T(), 216.66666666666666, stats.AverageExpense)

	assert.Equal(s.T(), 1000.0, stats.IncomeByCategory["salary"])
	assert.Equal(s.T(), 500.0, stats.IncomeByCategory["bonus"])
	assert.Len(s.T(), stats.IncomeByCategory, 2)

	assert.Equal(s.T(), 450.0, stats.ExpenseByCategory["food"])
	assert.Equal(s.T(), 200.0, stats.ExpenseByCategory["transport"])
	assert.Len(s.T(), stats.ExpenseByCategory, 2)

	assert.WithinDuration(s.T(), time.Now().UTC(), stats.GeneratedAt, time.Second)
}

func (s *CalculatorTestSuite) TestCalculateStats_OnlyIncome() {
	incomeTransactions := []domain.Transaction{
		{
			ID:        1,
			UserID:    "user-456",
			Amount:    2000.0,
			Type:      domain.TransactionTypeIncome,
			Category:  "salary",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			UserID:    "user-456",
			Amount:    1000.0,
			Type:      domain.TransactionTypeIncome,
			Category:  "investment",
			CreatedAt: time.Now(),
		},
	}

	stats := CalculateStats(incomeTransactions)

	assert.Equal(s.T(), "user-456", stats.UserID)
	assert.Equal(s.T(), 2, stats.TransactionsCount)
	assert.Equal(s.T(), 3000.0, stats.TotalIncome)
	assert.Equal(s.T(), 0.0, stats.TotalExpense)
	assert.Equal(s.T(), 3000.0, stats.Balance)
	assert.Equal(s.T(), 1500.0, stats.AverageIncome)
	assert.Equal(s.T(), 0.0, stats.AverageExpense)
	assert.Equal(s.T(), 2000.0, stats.IncomeByCategory["salary"])
	assert.Equal(s.T(), 1000.0, stats.IncomeByCategory["investment"])
	assert.Empty(s.T(), stats.ExpenseByCategory)
}

func (s *CalculatorTestSuite) TestCalculateStatsOnlyExpense() {
	expenseTransactions := []domain.Transaction{
		{
			ID:        1,
			UserID:    "user-789",
			Amount:    500.0,
			Type:      domain.TransactionTypeExpense,
			Category:  "rent",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			UserID:    "user-789",
			Amount:    300.0,
			Type:      domain.TransactionTypeExpense,
			Category:  "utilities",
			CreatedAt: time.Now(),
		},
	}

	stats := CalculateStats(expenseTransactions)

	assert.Equal(s.T(), "user-789", stats.UserID)
	assert.Equal(s.T(), 2, stats.TransactionsCount)
	assert.Equal(s.T(), 0.0, stats.TotalIncome)
	assert.Equal(s.T(), 800.0, stats.TotalExpense)
	assert.Equal(s.T(), -800.0, stats.Balance)
	assert.Equal(s.T(), 0.0, stats.AverageIncome)
	assert.Equal(s.T(), 400.0, stats.AverageExpense)
	assert.Equal(s.T(), 500.0, stats.ExpenseByCategory["rent"])
	assert.Equal(s.T(), 300.0, stats.ExpenseByCategory["utilities"])
	assert.Empty(s.T(), stats.IncomeByCategory)
}

func (s *CalculatorTestSuite) TestCalculateStatsSameCategoryMultipleTransactions() {
	transactions := []domain.Transaction{
		{
			ID:        1,
			UserID:    "user-999",
			Amount:    100.0,
			Type:      domain.TransactionTypeExpense,
			Category:  "food",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			UserID:    "user-999",
			Amount:    50.0,
			Type:      domain.TransactionTypeExpense,
			Category:  "food",
			CreatedAt: time.Now(),
		},
		{
			ID:        3,
			UserID:    "user-999",
			Amount:    30.0,
			Type:      domain.TransactionTypeExpense,
			Category:  "food",
			CreatedAt: time.Now(),
		},
	}

	stats := CalculateStats(transactions)

	assert.Equal(s.T(), 180.0, stats.ExpenseByCategory["food"])
	assert.Equal(s.T(), 1, len(stats.ExpenseByCategory))
	assert.Equal(s.T(), 60.0, stats.AverageExpense)
}

func (s *CalculatorTestSuite) TestCalculateStatsDifferentUsers() {
	transactions := []domain.Transaction{
		{
			ID:        1,
			UserID:    "user-first",
			Amount:    100.0,
			Type:      domain.TransactionTypeIncome,
			Category:  "salary",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			UserID:    "user-second",
			Amount:    200.0,
			Type:      domain.TransactionTypeIncome,
			Category:  "bonus",
			CreatedAt: time.Now(),
		},
	}

	stats := CalculateStats(transactions)

	assert.Equal(s.T(), "user-first", stats.UserID)
}

func (s *CalculatorTestSuite) TestCalculateStatsZeroAmounts() {
	transactions := []domain.Transaction{
		{
			ID:        1,
			UserID:    "user-zero",
			Amount:    0.0,
			Type:      domain.TransactionTypeIncome,
			Category:  "gift",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			UserID:    "user-zero",
			Amount:    0.0,
			Type:      domain.TransactionTypeExpense,
			Category:  "food",
			CreatedAt: time.Now(),
		},
	}

	stats := CalculateStats(transactions)

	assert.Equal(s.T(), 0.0, stats.TotalIncome)
	assert.Equal(s.T(), 0.0, stats.TotalExpense)
	assert.Equal(s.T(), 0.0, stats.Balance)
	assert.Equal(s.T(), 0.0, stats.AverageIncome)
	assert.Equal(s.T(), 0.0, stats.AverageExpense)
	assert.Equal(s.T(), 0.0, stats.IncomeByCategory["gift"])
	assert.Equal(s.T(), 0.0, stats.ExpenseByCategory["food"])
}

func TestCalculatorTestSuite(t *testing.T) {
	suite.Run(t, new(CalculatorTestSuite))
}
