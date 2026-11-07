package transaction

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func fromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	t := nt.Time
	return &t
}

func toNullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func fromNullInt64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	v := ni.Int64
	return &v
}

func toStudentPaymentDetailsResponse(studentID uuid.UUID, paidUntil sql.NullTime, sessionsAdded sql.NullInt64, studentName *string) *StudentPaymentDetailsResponse {
	return &StudentPaymentDetailsResponse{
		DetailsType:   DetailsTypeStudent,
		StudentID:     studentID,
		PaidUntil:     fromNullTime(paidUntil),
		SessionsAdded: fromNullInt64(sessionsAdded),
		StudentName:   studentName,
	}
}

func toStudentBalanceFromOwnerRow(b db.FindStudentBalancesByOwnerIDRow) StudentBalance {
	return StudentBalance{
		StudentID:          b.StudentID,
		PaidUntil:          fromNullTime(b.PaidUntil),
		SessionsLeft:       b.SessionsLeft,
		TotalSessionsAdded: b.TotalSessionsAdded,
	}
}

func toStudentBalanceFromIDsRow(b db.FindStudentBalancesByStudentIDsAndOwnerIDRow) StudentBalance {
	return StudentBalance{
		StudentID:          b.StudentID,
		PaidUntil:          fromNullTime(b.PaidUntil),
		SessionsLeft:       b.SessionsLeft,
		TotalSessionsAdded: b.TotalSessionsAdded,
	}
}

func toStudentBalance(b db.PaymentsModuleStudentBalance) StudentBalance {
	return StudentBalance{
		StudentID:    b.StudentID,
		PaidUntil:    fromNullTime(b.PaidUntil),
		SessionsLeft: b.SessionsLeft,
	}
}

func toTransactionResponse(tx db.PaymentsModuleTransaction, details *StudentPaymentDetailsResponse) TransactionResponse {
	amt, _ := strconv.ParseFloat(tx.Amount, 64)
	return TransactionResponse{
		ID:             tx.ID,
		Amount:         amt,
		Date:           tx.Date,
		Type:           TransactionType(tx.Type),
		CreatedAt:      tx.CreatedAt,
		PaymentDetails: details,
	}
}

func toTransactionResponseFromPaginatedRow(r db.FindTransactionsByOwnerIDPaginatedRow, details *StudentPaymentDetailsResponse) TransactionResponse {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	return TransactionResponse{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: details,
	}
}

func toTransactionResponseFromDateBetweenRow(r db.FindTransactionsByOwnerIDAndDateBetweenRow) TransactionResponse {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	var details *StudentPaymentDetailsResponse
	if r.StudentID.Valid && (r.PaidUntil.Valid || r.SessionsAdded.Valid) {
		details = toStudentPaymentDetailsResponse(r.StudentID.UUID, r.PaidUntil, r.SessionsAdded, nil)
	}
	return TransactionResponse{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: details,
	}
}

func toTransactionResponseFromStudentPaymentRow(r db.FindStudentPaymentsByStudentIDAndOwnerIDRow) TransactionResponse {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	return TransactionResponse{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: toStudentPaymentDetailsResponse(r.StudentID, r.PaidUntil, r.SessionsAdded, nil),
	}
}

func toTransactionResponseFromStudentPaymentsRow(r db.FindStudentPaymentsByStudentIDsAndOwnerIDRow) TransactionResponse {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	return TransactionResponse{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: toStudentPaymentDetailsResponse(r.StudentID, r.PaidUntil, r.SessionsAdded, nil),
	}
}
