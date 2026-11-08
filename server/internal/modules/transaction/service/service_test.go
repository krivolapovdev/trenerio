package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) {
	// RFC3339
	t1, err := parseTime("2026-10-17T15:04:05Z")
	require.NoError(t, err)
	assert.Equal(t, 2026, t1.Year())

	// RFC3339Nano
	t2, err := parseTime("2026-10-17T15:04:05.123456789Z")
	require.NoError(t, err)
	assert.Equal(t, 2026, t2.Year())

	// DateOnly
	t3, err := parseTime("2026-10-17")
	require.NoError(t, err)
	assert.Equal(t, 17, t3.Day())

	// time.Time type
	now := time.Now()
	t4, err := parseTime(now)
	require.NoError(t, err)
	assert.Equal(t, now, t4)

	// Invalid string
	_, err = parseTime("invalid-date")
	require.Error(t, err)

	// Unsupported type
	_, err = parseTime(12345)
	require.Error(t, err)
}

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
		_, err := svc.CreateTransaction(ctx, request.CreateTransactionRequest{
			Amount: 1000,
			Date:   pastDate,
			Type:   request.TransactionTypeIncome,
			PaymentDetails: &request.CreateStudentPaymentDetailsRequest{
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
			"sessionsAdded": 2.5,
		})
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Contains(t, badReq.Message, "positive integer")
	})

	t.Run("UpdateTransaction_NonPositiveSessionsAdded", func(t *testing.T) {
		_, err := svc.UpdateTransaction(ctx, uuid.New(), ownerID, map[string]any{
			"sessionsAdded": float64(0),
		})
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Contains(t, badReq.Message, "greater than 0")
	})
}
