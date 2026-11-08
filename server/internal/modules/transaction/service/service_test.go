package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPaginatedTransactions_InvalidPageSize(t *testing.T) {
	svc := NewService(nil, nil, nil)
	ctx := context.Background()

	_, err := svc.GetPaginatedTransactions(ctx, 1, 0, uuid.New())
	require.Error(t, err)
	var badReq *pkgerrors.BadRequestError
	require.ErrorAs(t, err, &badReq)
	assert.Contains(t, badReq.Message, "size must be at least 1")
}

func TestTransactionService_HybridValidation(t *testing.T) {
	svc := NewService(nil, nil, nil)
	ctx := context.Background()
	ownerID := uuid.New()
	studentID := uuid.New()
	pastDate := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)

	t.Run("CreateTransaction_MissingBothPaidUntilAndSessionsAdded", func(t *testing.T) {
		_, err := svc.CreateTransaction(ctx, domain.Transaction{
			Amount: 1000,
			Date:   pastDate,
			Type:   domain.TransactionTypeIncome,
			PaymentDetails: &domain.StudentPaymentDetails{
				StudentID: studentID,
			},
		}, ownerID)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Contains(t, badReq.Message, "either paidUntil or sessionsAdded")
	})

	t.Run("CreateStudentPayment_ZeroSessionsAdded", func(t *testing.T) {
		zero := int64(0)
		_, err := svc.CreateStudentPayment(ctx, studentID, 1000, pastDate, nil, &zero, ownerID)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Contains(t, badReq.Message, "sessionsAdded must be greater than 0")
	})

	t.Run("UpdateTransaction_InvalidSessionsAdded", func(t *testing.T) {
		_, err := svc.UpdateTransaction(ctx, uuid.New(), ownerID, map[string]any{
			"paymentDetails": map[string]any{
				"sessionsAdded": 2.5,
			},
		})
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Contains(t, badReq.Message, "positive integer")
	})

	t.Run("UpdateTransaction_NonPositiveSessionsAdded", func(t *testing.T) {
		_, err := svc.UpdateTransaction(ctx, uuid.New(), ownerID, map[string]any{
			"paymentDetails": map[string]any{
				"sessionsAdded": float64(0),
			},
		})
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Contains(t, badReq.Message, "greater than 0")
	})
}
