package mapper

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToTransactionResponse(t *testing.T) {
	txID := uuid.New()
	now := time.Now()
	date := time.Date(2026, 10, 20, 14, 0, 0, 0, time.UTC)

	tx := db.PaymentsModuleTransaction{
		ID:        txID,
		Amount:    "125.75",
		Date:      date,
		Type:      db.PaymentsModuleTransactionTypeINCOME,
		CreatedAt: now,
	}

	studentID := uuid.New()
	paidUntil := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	sessions := int64(8)
	details := &response.StudentPaymentDetailsResponse{
		DetailsType:   model.DetailsTypeStudent,
		StudentID:     studentID,
		PaidUntil:     &paidUntil,
		SessionsAdded: &sessions,
	}

	resp := ToTransactionResponse(tx, details)
	assert.Equal(t, txID, resp.ID)
	assert.Equal(t, 125.75, resp.Amount)
	assert.Equal(t, date, resp.Date)
	assert.Equal(t, model.TransactionTypeIncome, resp.Type)
	assert.Equal(t, now, resp.CreatedAt)
	assert.Equal(t, details, resp.PaymentDetails)
}

func TestStudentPaymentMappers_HybridModes(t *testing.T) {
	studentID := uuid.New()
	paidUntil := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

	t.Run("SessionOnlyRow", func(t *testing.T) {
		row := db.FindStudentPaymentsByStudentIDAndOwnerIDRow{
			ID:            uuid.New(),
			Amount:        "2000.00",
			Date:          time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			Type:          db.PaymentsModuleTransactionTypeINCOME,
			CreatedAt:     time.Now(),
			StudentID:     studentID,
			PaidUntil:     sql.NullTime{Valid: false},
			SessionsAdded: sql.NullInt64{Int64: 10, Valid: true},
		}

		resp := ToTransactionResponseFromStudentPaymentRow(row)
		require.NotNil(t, resp.PaymentDetails)
		assert.Nil(t, resp.PaymentDetails.PaidUntil)
		require.NotNil(t, resp.PaymentDetails.SessionsAdded)
		assert.Equal(t, int64(10), *resp.PaymentDetails.SessionsAdded)
	})

	t.Run("StudentBalanceMapper", func(t *testing.T) {
		bal := ToStudentBalance(db.PaymentsModuleStudentBalance{
			StudentID:    studentID,
			PaidUntil:    sql.NullTime{Time: paidUntil, Valid: true},
			SessionsLeft: 4,
		})
		assert.Equal(t, studentID, bal.StudentID)
		require.NotNil(t, bal.PaidUntil)
		assert.Equal(t, paidUntil, *bal.PaidUntil)
		assert.Equal(t, int64(4), bal.SessionsLeft)
	})
}
