package http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/harness"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
)

type mockTxService struct {
	getPaginatedTransactionsFn func(ctx context.Context, page, size int, ownerID uuid.UUID) (domain.PageTransaction, error)
	createTransactionFn        func(ctx context.Context, txToCreate domain.Transaction, ownerID uuid.UUID) (domain.Transaction, error)
	updateTransactionFn        func(ctx context.Context, txID, ownerID uuid.UUID, updates map[string]any) (domain.Transaction, error)
	deleteTransactionFn        func(ctx context.Context, txID, ownerID uuid.UUID) error
}

func (m *mockTxService) GetPaginatedTransactions(ctx context.Context, page, size int, ownerID uuid.UUID) (domain.PageTransaction, error) {
	if m.getPaginatedTransactionsFn != nil {
		return m.getPaginatedTransactionsFn(ctx, page, size, ownerID)
	}
	return domain.PageTransaction{}, nil
}

func (m *mockTxService) GetTransactionsBetween() ([]domain.Transaction, error) {
	return nil, nil
}

func (m *mockTxService) GetTransactionsByStudentID() ([]domain.Transaction, error) {
	return nil, nil
}

func (m *mockTxService) GetTransactionsByStudentIDs() (map[uuid.UUID][]domain.Transaction, error) {
	return nil, nil
}

func (m *mockTxService) CreateTransaction(ctx context.Context, txToCreate domain.Transaction, ownerID uuid.UUID) (domain.Transaction, error) {
	if m.createTransactionFn != nil {
		return m.createTransactionFn(ctx, txToCreate, ownerID)
	}
	return domain.Transaction{}, nil
}

func (m *mockTxService) CreateStudentPayment() (domain.Transaction, error) {
	return domain.Transaction{}, nil
}

func (m *mockTxService) UpdateTransaction(ctx context.Context, transactionID, ownerID uuid.UUID, updates map[string]any) (domain.Transaction, error) {
	if m.updateTransactionFn != nil {
		return m.updateTransactionFn(ctx, transactionID, ownerID, updates)
	}
	return domain.Transaction{}, nil
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
