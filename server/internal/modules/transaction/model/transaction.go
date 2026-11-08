package model

import "github.com/krivolapovdev/trenerio/internal/modules/transaction/request"

type TransactionType = request.TransactionType

const (
	TransactionTypeIncome  = request.TransactionTypeIncome
	TransactionTypeExpense = request.TransactionTypeExpense
)

const (
	DetailsTypeStudent = "STUDENT"
)
