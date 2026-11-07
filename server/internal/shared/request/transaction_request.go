package request

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "INCOME"
	TransactionTypeExpense TransactionType = "EXPENSE"
)

// @name CreateStudentPaymentDetailsRequest
type CreateStudentPaymentDetailsRequest struct {
	DetailsType   string     `json:"detailsType,omitempty"`
	StudentID     uuid.UUID  `json:"studentId" binding:"required" format:"uuid"`
	PaidUntil     *time.Time `json:"paidUntil,omitempty" format:"date-time"`
	SessionsAdded *int64     `json:"sessionsAdded,omitempty" binding:"omitempty,gt=0"`
}

// @name CreateTransactionRequest
type CreateTransactionRequest struct {
	Amount         float64                             `json:"amount" binding:"required,gt=0"`
	Date           time.Time                           `json:"date" binding:"required" format:"date-time"`
	Type           TransactionType                     `json:"type" binding:"required"`
	PaymentDetails *CreateStudentPaymentDetailsRequest `json:"paymentDetails,omitempty"`
}
