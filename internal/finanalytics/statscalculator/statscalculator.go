package statscalculator

import (
	"time"

	"fin-track-app/internal/domain"
)

func CalculateStats(transactions []domain.Transaction) domain.FinanceStats {
	stats := domain.FinanceStats{
		ExpenseByCategory: map[string]float64{},
		IncomeByCategory:  map[string]float64{},
		TransactionsCount: len(transactions),
		GeneratedAt:       time.Now().UTC(),
	}

	var incomeSamples []float64
	var expenseSamples []float64

	for _, tx := range transactions {
		if stats.UserID == 0 {
			stats.UserID = tx.UserID
		}

		switch tx.Type {
		case domain.TransactionTypeIncome:
			stats.TotalIncome += tx.Amount
			incomeSamples = append(incomeSamples, tx.Amount)
			stats.IncomeByCategory[tx.Category] += tx.Amount
		case domain.TransactionTypeExpense:
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
