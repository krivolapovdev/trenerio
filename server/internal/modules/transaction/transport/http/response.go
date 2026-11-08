package http

import (
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
)

// @name StudentPaymentDetailsResponse
type StudentPaymentDetailsResponse struct {
	DetailsType   string     `json:"detailsType" binding:"required"`
	StudentID     uuid.UUID  `json:"studentId" binding:"required" format:"uuid"`
	PaidFrom      *time.Time `json:"paidFrom,omitempty" format:"date-time"`
	PaidUntil     *time.Time `json:"paidUntil,omitempty" format:"date-time"`
	SessionsAdded *int64     `json:"sessionsAdded,omitempty"`
	StudentName   *string    `json:"studentName,omitempty"`
}

// @name TransactionResponse
type TransactionResponse struct {
	ID             uuid.UUID                      `json:"id" binding:"required" format:"uuid"`
	Amount         float64                        `json:"amount" binding:"required"`
	Date           time.Time                      `json:"date" binding:"required" format:"date-time"`
	Type           domain.TransactionType         `json:"type" binding:"required"`
	CreatedAt      time.Time                      `json:"createdAt" binding:"required" format:"date-time"`
	PaymentDetails *StudentPaymentDetailsResponse `json:"paymentDetails,omitempty"`
}

// @name SortObject
type SortObject struct {
	Empty    bool `json:"empty"`
	Sorted   bool `json:"sorted"`
	Unsorted bool `json:"unsorted"`
}

// @name PageableObject
type PageableObject struct {
	Offset     int64       `json:"offset"`
	PageNumber int         `json:"pageNumber"`
	PageSize   int         `json:"pageSize"`
	Paged      bool        `json:"paged"`
	Unpaged    bool        `json:"unpaged"`
	Sort       *SortObject `json:"sort,omitempty"`
}

// @name PageTransactionResponse
type PageTransactionResponse struct {
	TotalPages       int                   `json:"totalPages"`
	TotalElements    int64                 `json:"totalElements"`
	Size             int                   `json:"size"`
	Content          []TransactionResponse `json:"content"`
	Number           int                   `json:"number"`
	First            bool                  `json:"first"`
	Last             bool                  `json:"last"`
	NumberOfElements int                   `json:"numberOfElements"`
	Empty            bool                  `json:"empty"`
	Sort             *SortObject           `json:"sort,omitempty"`
	Pageable         *PageableObject       `json:"pageable,omitempty"`
}

func ToStudentPaymentDetailsResponse(d *domain.StudentPaymentDetails) *StudentPaymentDetailsResponse {
	if d == nil {
		return nil
	}
	return &StudentPaymentDetailsResponse{
		DetailsType:   d.DetailsType,
		StudentID:     d.StudentID,
		PaidFrom:      d.PaidFrom,
		PaidUntil:     d.PaidUntil,
		SessionsAdded: d.SessionsAdded,
		StudentName:   d.StudentName,
	}
}

func ToTransactionResponse(tx domain.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:             tx.ID,
		Amount:         tx.Amount,
		Date:           tx.Date,
		Type:           tx.Type,
		CreatedAt:      tx.CreatedAt,
		PaymentDetails: ToStudentPaymentDetailsResponse(tx.PaymentDetails),
	}
}
