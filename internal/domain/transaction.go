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
	IncomeByCategory  map[string]float64 `json:"income_by_category"`
	TransactionsCount int                `json:"transactions_count"`
	GeneratedAt       time.Time          `json:"generated_at"`
}
