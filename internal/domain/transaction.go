package domain

import (
	"time"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

type Transaction struct {
	ID        int64           `json:"id"`
	UserID    string          `json:"user_id"`
	Amount    float64         `json:"amount"`
	Category  string          `json:"category"`
	Type      TransactionType `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
}

type TransactionMessage struct {
	UserID       string        `json:"user_id"`
	Transactions []Transaction `json:"transactions"`
}

type FinanceStats struct {
	UserID            string             `json:"user_id"`
	TotalIncome       float64            `json:"total_income"`
	TotalExpense      float64            `json:"total_expense"`
	Balance           float64            `json:"balance"`
	AverageIncome     float64            `json:"average_income"`
	AverageExpense    float64            `json:"average_expense"`
	ExpenseByCategory map[string]float64 `json:"expense_by_category"`
	IncomeByCategory map[string]float64 `json:"income_by_category"`
	TransactionsCount int                `json:"transactions_count"`
	GeneratedAt       time.Time          `json:"generated_at"`
}

func CalculateStats(transactions []Transaction) FinanceStats {
	stats := FinanceStats{
		ExpenseByCategory: map[string]float64{},
		IncomeByCategory: map[string]float64{},
		TransactionsCount: len(transactions),
		GeneratedAt:       time.Now().UTC(),
	}

	var incomeSamples []float64
	var expenseSamples []float64

	for _, tx := range transactions {
		if stats.UserID == "" {
			stats.UserID = tx.UserID
		}

		switch tx.Type {
		case TransactionTypeIncome:
			stats.TotalIncome += tx.Amount
			incomeSamples = append(incomeSamples, tx.Amount)
			stats.IncomeByCategory[tx.Category] += tx.Amount
		case TransactionTypeExpense:
			stats.TotalExpense += tx.Amount
			expenseSamples = append(expenseSamples, tx.Amount)
			stats.ExpenseByCategory[tx.Category] += tx.Amount
		}
	}

	stats.Balance = stats.TotalIncome - stats.TotalExpense
	stats.AverageIncome = average(incomeSamples)
	stats.AverageExpense = average(expenseSamples)

	return stats
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
