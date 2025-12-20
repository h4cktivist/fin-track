package domain

import "errors"

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
)
