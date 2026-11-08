package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionHandler_GetPaginatedTransactions(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupTxRouter(&mockTxService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/transactions", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("SuccessWithDefaults", func(t *testing.T) {
		txID := uuid.New()
		studentID := uuid.New()
		studentName := "John Doe"
		mockSvc := &mockTxService{
			getPaginatedTransactionsFn: func(ctx context.Context, page, size int, ownerID uuid.UUID) (txmodel.PageTransaction, error) {
				assert.Equal(t, 1, page)
				assert.Equal(t, 20, size)
				assert.Equal(t, testUser.ID, ownerID)
				return txmodel.PageTransaction{
					Number:           0,
					Size:             size,
					TotalElements:    25,
					TotalPages:       2,
					First:            true,
					Last:             false,
					NumberOfElements: 1,
					Empty:            false,
					Content: []txmodel.Transaction{
						{
							ID:     txID,
							Amount: 1500,
							Type:   txmodel.TransactionTypeIncome,
							PaymentDetails: &txmodel.StudentPaymentDetails{
								StudentID:   studentID,
								StudentName: &studentName,
							},
						},
					},
				}, nil
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/transactions", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
		var actual PageTransactionResponse
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		require.NoError(t, err)
		assert.Equal(t, 0, actual.Number)
		assert.Equal(t, 2, actual.TotalPages)
		assert.Equal(t, int64(25), actual.TotalElements)
		assert.True(t, actual.First)
		assert.False(t, actual.Last)
		assert.Len(t, actual.Content, 1)
		assert.Equal(t, txID, actual.Content[0].ID)
		assert.Equal(t, 1500.0, actual.Content[0].Amount)
		require.NotNil(t, actual.Content[0].PaymentDetails)
		assert.Equal(t, &studentName, actual.Content[0].PaymentDetails.StudentName)
		require.NotNil(t, actual.Pageable)
		assert.Equal(t, 20, actual.Pageable.PageSize)
		assert.True(t, actual.Pageable.Paged)
		require.NotNil(t, actual.Sort)
		assert.True(t, actual.Sort.Sorted)
	})

	t.Run("SuccessWithQueryParams", func(t *testing.T) {
		mockSvc := &mockTxService{
			getPaginatedTransactionsFn: func(ctx context.Context, page, size int, ownerID uuid.UUID) (txmodel.PageTransaction, error) {
				assert.Equal(t, 2, page)
				assert.Equal(t, 50, size)
				return txmodel.PageTransaction{Number: page, Size: size}, nil
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/transactions?page=2&size=50", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockTxService{
			getPaginatedTransactionsFn: func(ctx context.Context, page, size int, ownerID uuid.UUID) (txmodel.PageTransaction, error) {
				return txmodel.PageTransaction{}, assert.AnError
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/transactions", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}
