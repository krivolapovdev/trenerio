package mapper

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/model"
)

func ToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func FromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time
	return &t
}

func ToNullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func FromNullInt64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	v := ni.Int64
	return &v
}

func ToStudentPaymentDetailsResponse(studentID uuid.UUID, paidUntil sql.NullTime, sessionsAdded sql.NullInt64, studentName *string) *response.StudentPaymentDetailsResponse {
	return &response.StudentPaymentDetailsResponse{
		DetailsType:   model.DetailsTypeStudent,
		StudentID:     studentID,
		PaidUntil:     FromNullTime(paidUntil),
		SessionsAdded: FromNullInt64(sessionsAdded),
		StudentName:   studentName,
	}
}

func ToStudentBalanceFromOwnerRow(b db.FindStudentBalancesByOwnerIDRow) model.StudentBalance {
	return model.StudentBalance{
		StudentID:          b.StudentID,
		PaidUntil:          FromNullTime(b.PaidUntil),
		SessionsLeft:       b.SessionsLeft,
		TotalSessionsAdded: b.TotalSessionsAdded,
	}
}

func ToStudentBalanceFromIDsRow(b db.FindStudentBalancesByStudentIDsAndOwnerIDRow) model.StudentBalance {
	return model.StudentBalance{
		StudentID:          b.StudentID,
		PaidUntil:          FromNullTime(b.PaidUntil),
		SessionsLeft:       b.SessionsLeft,
		TotalSessionsAdded: b.TotalSessionsAdded,
	}
}

func ToStudentBalance(b db.PaymentsModuleStudentBalance) model.StudentBalance {
	return model.StudentBalance{
		StudentID:    b.StudentID,
		PaidUntil:    FromNullTime(b.PaidUntil),
		SessionsLeft: b.SessionsLeft,
	}
}

func ToTransactionResponse(tx db.PaymentsModuleTransaction, details *response.StudentPaymentDetailsResponse) response.TransactionResponse {
	amt, _ := strconv.ParseFloat(tx.Amount, 64)
	return response.TransactionResponse{
		ID:             tx.ID,
		Amount:         amt,
		Date:           tx.Date,
		Type:           model.TransactionType(tx.Type),
		CreatedAt:      tx.CreatedAt,
		PaymentDetails: details,
	}
}

func ToTransactionResponseFromPaginatedRow(r db.FindTransactionsByOwnerIDPaginatedRow, details *response.StudentPaymentDetailsResponse) response.TransactionResponse {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	return response.TransactionResponse{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           model.TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: details,
	}
}

func ToTransactionResponseFromDateBetweenRow(r db.FindTransactionsByOwnerIDAndDateBetweenRow) response.TransactionResponse {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	var details *response.StudentPaymentDetailsResponse
	if r.StudentID.Valid && (r.PaidUntil.Valid || r.SessionsAdded.Valid) {
		details = ToStudentPaymentDetailsResponse(r.StudentID.UUID, r.PaidUntil, r.SessionsAdded, nil)
	}
	return response.TransactionResponse{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           model.TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: details,
	}
}

func ToTransactionResponseFromStudentPaymentRow(r db.FindStudentPaymentsByStudentIDAndOwnerIDRow) response.TransactionResponse {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	return response.TransactionResponse{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           model.TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: ToStudentPaymentDetailsResponse(r.StudentID, r.PaidUntil, r.SessionsAdded, nil),
	}
}

func ToTransactionResponseFromStudentPaymentsRow(r db.FindStudentPaymentsByStudentIDsAndOwnerIDRow) response.TransactionResponse {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	return response.TransactionResponse{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           model.TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: ToStudentPaymentDetailsResponse(r.StudentID, r.PaidUntil, r.SessionsAdded, nil),
	}
}
