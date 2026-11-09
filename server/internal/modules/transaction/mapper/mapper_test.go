package mapper

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToTransaction(t *testing.T) {
	txID := uuid.New()
	now := time.Now()
	date := time.Date(2026, 10, 20, 14, 0, 0, 0, time.UTC)

	tx := db.FinanceModuleTransaction{
		ID:        txID,
		Amount:    "125.75",
		Date:      date,
		Type:      db.FinanceModuleTransactionTypeINCOME,
		CreatedAt: now,
	}

	studentID := uuid.New()
	paidUntil := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	sessions := int64(8)
	details := &domain.StudentPaymentDetails{
		DetailsType:   domain.DetailsTypeStudent,
		StudentID:     studentID,
		PaidUntil:     &paidUntil,
		SessionsAdded: &sessions,
	}

	res := ToTransaction(tx, details)
	assert.Equal(t, txID, res.ID)
	assert.Equal(t, 125.75, res.Amount)
	assert.Equal(t, date, res.Date)
	assert.Equal(t, domain.TransactionTypeIncome, res.Type)
	assert.Equal(t, now, res.CreatedAt)
	assert.Equal(t, details, res.PaymentDetails)
}

func TestStudentPaymentMappers_HybridModes(t *testing.T) {
	studentID := uuid.New()
	paidFrom := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	paidUntil := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	sessions := int64(10)
	now := time.Now()

	singleRow := db.FindStudentPaymentsByStudentIDAndOwnerIDRow{
		ID:            uuid.New(),
		Amount:        "5000",
		Date:          now,
		Type:          db.FinanceModuleTransactionTypeINCOME,
		CreatedAt:     now,
		StudentID:     studentID,
		PaidFrom:      sql.NullTime{Time: paidFrom, Valid: true},
		PaidUntil:     sql.NullTime{Time: paidUntil, Valid: true},
		SessionsAdded: sql.NullInt64{Int64: sessions, Valid: true},
	}

	singleTx := ToTransactionFromStudentPaymentRow(singleRow)
	require.NotNil(t, singleTx.PaymentDetails)
	assert.Equal(t, &paidFrom, singleTx.PaymentDetails.PaidFrom)
	assert.Equal(t, &paidUntil, singleTx.PaymentDetails.PaidUntil)
	assert.Equal(t, &sessions, singleTx.PaymentDetails.SessionsAdded)

	multiRow := db.FindStudentPaymentsByStudentIDsAndOwnerIDRow{
		ID:            uuid.New(),
		Amount:        "3000",
		Date:          now,
		Type:          db.FinanceModuleTransactionTypeINCOME,
		CreatedAt:     now,
		StudentID:     studentID,
		PaidFrom:      sql.NullTime{Time: paidFrom, Valid: true},
		PaidUntil:     sql.NullTime{Time: paidUntil, Valid: true},
		SessionsAdded: sql.NullInt64{Int64: sessions, Valid: true},
	}

	multiTx := ToTransactionFromStudentPaymentsRow(multiRow)
	require.NotNil(t, multiTx.PaymentDetails)
	assert.Equal(t, &paidFrom, multiTx.PaymentDetails.PaidFrom)
	assert.Equal(t, &paidUntil, multiTx.PaymentDetails.PaidUntil)
	assert.Equal(t, &sessions, multiTx.PaymentDetails.SessionsAdded)
}
