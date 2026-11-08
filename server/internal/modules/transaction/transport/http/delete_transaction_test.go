package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/model"
	"github.com/stretchr/testify/assert"
)

func TestTransactionHandler_DeleteTransaction(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	txID := uuid.New()

	t.Run("InvalidTxID", func(t *testing.T) {
		r := setupTxRouter(&mockTxService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/transactions/bad-id", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		called := false
		mockSvc := &mockTxService{
			deleteTransactionFn: func(ctx context.Context, id, ownerID uuid.UUID) error {
				assert.Equal(t, txID, id)
				assert.Equal(t, testUser.ID, ownerID)
				called = true
				return nil
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/transactions/"+txID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNoContent, w.Code)
		assert.True(t, called)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSvc := &mockTxService{
			deleteTransactionFn: func(ctx context.Context, id, ownerID uuid.UUID) error {
				return pkgerrors.NewNotFound("Transaction", txID, ownerID)
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/transactions/"+txID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockTxService{
			deleteTransactionFn: func(ctx context.Context, id, ownerID uuid.UUID) error {
				return assert.AnError
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/transactions/"+txID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}
