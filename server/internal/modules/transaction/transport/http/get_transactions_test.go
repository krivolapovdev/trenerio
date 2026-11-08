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
		mockSvc := &mockTxService{
			getPaginatedTransactionsFn: func(ctx context.Context, page, size int, ownerID uuid.UUID) (txmodel.PageTransaction, error) {
				assert.Equal(t, 1, page)
				assert.Equal(t, 20, size)
				assert.Equal(t, testUser.ID, ownerID)
				return txmodel.PageTransaction{
					Number:        page,
					Size:          size,
					TotalElements: 1,
					TotalPages:    1,
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
		assert.Equal(t, 1, actual.Number)
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
