package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "INCOME"
	TransactionTypeExpense TransactionType = "EXPENSE"
)

const (
	DetailsTypeStudent = "STUDENT"
)

type StudentPaymentDetails struct {
	DetailsType   string
	StudentID     uuid.UUID
	PaidFrom      *time.Time
	PaidUntil     *time.Time
	SessionsAdded *int64
	StudentName   *string
}

type Transaction struct {
	ID             uuid.UUID
	Amount         float64
	Date           time.Time
	Type           TransactionType
	CreatedAt      time.Time
	PaymentDetails *StudentPaymentDetails
}

type PageTransaction struct {
	TotalPages       int
	TotalElements    int64
	Size             int
	Content          []Transaction
	Number           int
	First            bool
	Last             bool
	NumberOfElements int
	Empty            bool
}
