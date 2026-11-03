package transaction

import "github.com/krivolapovdev/trenerio/internal/shared/request"

type TransactionType = request.TransactionType

const (
	TransactionTypeIncome  = request.TransactionTypeIncome
	TransactionTypeExpense = request.TransactionTypeExpense
)

const (
	DetailsTypeStudent = "STUDENT"
)
