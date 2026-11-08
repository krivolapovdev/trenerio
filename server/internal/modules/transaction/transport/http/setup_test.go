package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/harness"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/model"
)

type mockTxService struct {
	getPaginatedTransactionsFn func(ctx context.Context, page, size int, ownerID uuid.UUID) (PageTransactionResponse, error)
	createTransactionFn        func(ctx context.Context, req CreateTransactionRequest, ownerID uuid.UUID) (TransactionResponse, error)
	updateTransactionFn        func(ctx context.Context, txID, ownerID uuid.UUID, updates map[string]any) (TransactionResponse, error)
	deleteTransactionFn        func(ctx context.Context, txID, ownerID uuid.UUID) error
}

func (m *mockTxService) GetPaginatedTransactions(ctx context.Context, page, size int, ownerID uuid.UUID) (PageTransactionResponse, error) {
	if m.getPaginatedTransactionsFn != nil {
		return m.getPaginatedTransactionsFn(ctx, page, size, ownerID)
	}
	return PageTransactionResponse{}, nil
}

func (m *mockTxService) GetTransactionsBetween(ctx context.Context, startDate, endDate time.Time, ownerID uuid.UUID) ([]TransactionResponse, error) {
	return nil, nil
}

func (m *mockTxService) GetTransactionsByStudentID(ctx context.Context, studentID, ownerID uuid.UUID) ([]TransactionResponse, error) {
	return nil, nil
}

func (m *mockTxService) GetTransactionsByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID][]TransactionResponse, error) {
	return nil, nil
}

func (m *mockTxService) CreateTransaction(ctx context.Context, req CreateTransactionRequest, ownerID uuid.UUID) (TransactionResponse, error) {
	if m.createTransactionFn != nil {
		return m.createTransactionFn(ctx, req, ownerID)
	}
	return TransactionResponse{}, nil
}

func (m *mockTxService) CreateStudentPayment(ctx context.Context, studentID uuid.UUID, amount float64, date time.Time, paidUntil *time.Time, sessionsAdded *int64, ownerID uuid.UUID) (TransactionResponse, error) {
	return TransactionResponse{}, nil
}

func (m *mockTxService) UpdateTransaction(ctx context.Context, transactionID, ownerID uuid.UUID, updates map[string]any) (TransactionResponse, error) {
	if m.updateTransactionFn != nil {
		return m.updateTransactionFn(ctx, transactionID, ownerID, updates)
	}
	return TransactionResponse{}, nil
}

func (m *mockTxService) DeleteTransaction(ctx context.Context, transactionID, ownerID uuid.UUID) error {
	if m.deleteTransactionFn != nil {
		return m.deleteTransactionFn(ctx, transactionID, ownerID)
	}
	return nil
}

func setupTxRouter(svc transactionService, user authmodel.JwtUser) *gin.Engine {
	return harness.SetupTestRouter(user, func(rg *gin.RouterGroup) {
		NewHandler(svc).RegisterRoutes(rg)
	})
}
