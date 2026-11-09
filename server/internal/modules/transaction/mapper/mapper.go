package mapper

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
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

func ToStudentPaymentDetails(studentID uuid.UUID, paidFrom, paidUntil sql.NullTime, sessionsAdded sql.NullInt64, studentName *string) *domain.StudentPaymentDetails {
	return &domain.StudentPaymentDetails{
		DetailsType:   domain.DetailsTypeStudent,
		StudentID:     studentID,
		PaidFrom:      FromNullTime(paidFrom),
		PaidUntil:     FromNullTime(paidUntil),
		SessionsAdded: FromNullInt64(sessionsAdded),
		StudentName:   studentName,
	}
}

func ToStudentBalanceFromOwnerRow(b db.FindStudentBalancesByOwnerIDRow) domain.StudentBalance {
	return domain.StudentBalance{
		StudentID:          b.StudentID,
		PaidFrom:           FromNullTime(b.PaidFrom),
		PaidUntil:          FromNullTime(b.PaidUntil),
		SessionsLeft:       b.SessionsLeft,
		TotalSessionsAdded: b.TotalSessionsAdded,
	}
}

func ToStudentBalanceFromIDsRow(b db.FindStudentBalancesByStudentIDsAndOwnerIDRow) domain.StudentBalance {
	return domain.StudentBalance{
		StudentID:          b.StudentID,
		PaidFrom:           FromNullTime(b.PaidFrom),
		PaidUntil:          FromNullTime(b.PaidUntil),
		SessionsLeft:       b.SessionsLeft,
		TotalSessionsAdded: b.TotalSessionsAdded,
	}
}

func ToTransaction(tx db.FinanceModuleTransaction, details *domain.StudentPaymentDetails) domain.Transaction {
	amt, _ := strconv.ParseFloat(tx.Amount, 64)
	return domain.Transaction{
		ID:             tx.ID,
		Amount:         amt,
		Date:           tx.Date,
		Type:           domain.TransactionType(tx.Type),
		CreatedAt:      tx.CreatedAt,
		PaymentDetails: details,
	}
}

func ToTransactionFromPaginatedRow(r db.FindTransactionsByOwnerIDPaginatedRow, details *domain.StudentPaymentDetails) domain.Transaction {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	return domain.Transaction{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           domain.TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: details,
	}
}

func ToTransactionFromDateBetweenRow(r db.FindTransactionsByOwnerIDAndDateBetweenRow) domain.Transaction {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	var details *domain.StudentPaymentDetails
	if r.StudentID.Valid && (r.PaidUntil.Valid || r.PaidFrom.Valid || r.SessionsAdded.Valid) {
		details = ToStudentPaymentDetails(r.StudentID.UUID, r.PaidFrom, r.PaidUntil, r.SessionsAdded, nil)
	}
	return domain.Transaction{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           domain.TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: details,
	}
}

func ToTransactionFromStudentPaymentRow(r db.FindStudentPaymentsByStudentIDAndOwnerIDRow) domain.Transaction {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	return domain.Transaction{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           domain.TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: ToStudentPaymentDetails(r.StudentID, r.PaidFrom, r.PaidUntil, r.SessionsAdded, nil),
	}
}

func ToTransactionFromStudentPaymentsRow(r db.FindStudentPaymentsByStudentIDsAndOwnerIDRow) domain.Transaction {
	amt, _ := strconv.ParseFloat(r.Amount, 64)
	return domain.Transaction{
		ID:             r.ID,
		Amount:         amt,
		Date:           r.Date,
		Type:           domain.TransactionType(r.Type),
		CreatedAt:      r.CreatedAt,
		PaymentDetails: ToStudentPaymentDetails(r.StudentID, r.PaidFrom, r.PaidUntil, r.SessionsAdded, nil),
	}
}
