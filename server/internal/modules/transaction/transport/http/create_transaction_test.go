package http

import (
	"bytes"
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionHandler_CreateTransaction(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupTxRouter(&mockTxService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/transactions", bytes.NewBufferString(`{"amount":1000}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		r := setupTxRouter(&mockTxService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/transactions", bytes.NewBufferString(`{invalid`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		txDate := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
		expectedID := uuid.New()

		mockSvc := &mockTxService{
			createTransactionFn: func(ctx context.Context, txToCreate txmodel.Transaction, ownerID uuid.UUID) (txmodel.Transaction, error) {
				assert.Equal(t, testUser.ID, ownerID)
				assert.Equal(t, 2500.0, txToCreate.Amount)
				return txmodel.Transaction{ID: expectedID, Amount: txToCreate.Amount}, nil
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		payload := `{"amount":2500.0,"type":"INCOME","date":"` + txDate.Format(time.RFC3339) + `"}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/transactions", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusCreated, w.Code)
		var actual TransactionResponse
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		require.NoError(t, err)
		assert.Equal(t, expectedID, actual.ID)
	})

	t.Run("Success_SessionOnlyStudentPayment", func(t *testing.T) {
		txDate := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
		studentID := uuid.New()
		expectedID := uuid.New()

		mockSvc := &mockTxService{
			createTransactionFn: func(ctx context.Context, txToCreate txmodel.Transaction, ownerID uuid.UUID) (txmodel.Transaction, error) {
				require.NotNil(t, txToCreate.PaymentDetails)
				assert.Equal(t, studentID, txToCreate.PaymentDetails.StudentID)
				assert.Nil(t, txToCreate.PaymentDetails.PaidUntil)
				require.NotNil(t, txToCreate.PaymentDetails.SessionsAdded)
				assert.Equal(t, int64(8), *txToCreate.PaymentDetails.SessionsAdded)
				return txmodel.Transaction{
					ID:     expectedID,
					Amount: txToCreate.Amount,
					PaymentDetails: &txmodel.StudentPaymentDetails{
						StudentID:     studentID,
						SessionsAdded: txToCreate.PaymentDetails.SessionsAdded,
					},
				}, nil
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		payload := `{"amount":2500.0,"type":"INCOME","date":"` + txDate.Format(time.RFC3339) + `","paymentDetails":{"studentId":"` + studentID.String() + `","sessionsAdded":8}}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/transactions", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusCreated, w.Code)
	})

	t.Run("Success_HybridStudentPayment", func(t *testing.T) {
		txDate := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
		paidUntil := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
		studentID := uuid.New()
		expectedID := uuid.New()

		mockSvc := &mockTxService{
			createTransactionFn: func(ctx context.Context, txToCreate txmodel.Transaction, ownerID uuid.UUID) (txmodel.Transaction, error) {
				require.NotNil(t, txToCreate.PaymentDetails)
				require.NotNil(t, txToCreate.PaymentDetails.PaidUntil)
				require.NotNil(t, txToCreate.PaymentDetails.SessionsAdded)
				assert.Equal(t, int64(4), *txToCreate.PaymentDetails.SessionsAdded)
				return txmodel.Transaction{
					ID:     expectedID,
					Amount: txToCreate.Amount,
					PaymentDetails: &txmodel.StudentPaymentDetails{
						StudentID:     studentID,
						PaidUntil:     txToCreate.PaymentDetails.PaidUntil,
						SessionsAdded: txToCreate.PaymentDetails.SessionsAdded,
					},
				}, nil
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		payload := `{"amount":5000.0,"type":"INCOME","date":"` + txDate.Format(time.RFC3339) + `","paymentDetails":{"studentId":"` + studentID.String() + `","paidUntil":"` + paidUntil.Format(time.RFC3339) + `","sessionsAdded":4}}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/transactions", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusCreated, w.Code)
	})

	t.Run("BadRequest_MissingPaidUntilAndSessionsAdded", func(t *testing.T) {
		studentID := uuid.New()
		r := setupTxRouter(&mockTxService{}, testUser)
		w := httptest.NewRecorder()
		payload := `{"amount":2500.0,"type":"INCOME","date":"2026-10-17T10:00:00Z","paymentDetails":{"studentId":"` + studentID.String() + `"}}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/transactions", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("BadRequest_ZeroSessionsAdded", func(t *testing.T) {
		studentID := uuid.New()
		r := setupTxRouter(&mockTxService{}, testUser)
		w := httptest.NewRecorder()
		payload := `{"amount":2500.0,"type":"INCOME","date":"2026-10-17T10:00:00Z","paymentDetails":{"studentId":"` + studentID.String() + `","sessionsAdded":0}}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/transactions", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockTxService{
			createTransactionFn: func(ctx context.Context, txToCreate txmodel.Transaction, ownerID uuid.UUID) (txmodel.Transaction, error) {
				return txmodel.Transaction{}, assert.AnError
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		payload := `{"amount":2500.0,"type":"INCOME","date":"2026-10-17T10:00:00Z"}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/transactions", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}
