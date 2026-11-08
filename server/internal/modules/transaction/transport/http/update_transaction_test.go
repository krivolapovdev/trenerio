package http

import (
	"bytes"
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/stretchr/testify/assert"
)

func TestTransactionHandler_UpdateTransaction(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	txID := uuid.New()

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupTxRouter(&mockTxService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/transactions/"+txID.String(), bytes.NewBufferString(`{"amount":3000.0}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidTxID", func(t *testing.T) {
		r := setupTxRouter(&mockTxService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/transactions/bad-id", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		r := setupTxRouter(&mockTxService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/transactions/"+txID.String(), bytes.NewBufferString(`{bad`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		mockSvc := &mockTxService{
			updateTransactionFn: func(ctx context.Context, id, ownerID uuid.UUID, updates map[string]any) (txmodel.Transaction, error) {
				assert.Equal(t, txID, id)
				assert.Equal(t, testUser.ID, ownerID)
				assert.Equal(t, 3000.0, updates["amount"])
				return txmodel.Transaction{ID: txID, Amount: 3000.0}, nil
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/transactions/"+txID.String(), bytes.NewBufferString(`{"amount":3000.0}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSvc := &mockTxService{
			updateTransactionFn: func(ctx context.Context, id, ownerID uuid.UUID, updates map[string]any) (txmodel.Transaction, error) {
				return txmodel.Transaction{}, pkgerrors.NewNotFound("Transaction", id, ownerID)
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/transactions/"+txID.String(), bytes.NewBufferString(`{"amount":3000.0}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockTxService{
			updateTransactionFn: func(ctx context.Context, id, ownerID uuid.UUID, updates map[string]any) (txmodel.Transaction, error) {
				return txmodel.Transaction{}, assert.AnError
			},
		}

		r := setupTxRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/transactions/"+txID.String(), bytes.NewBufferString(`{"amount":3000.0}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}
