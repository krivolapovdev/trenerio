package http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
)

type transactionService interface {
	GetPaginatedTransactions(ctx context.Context, page, size int, ownerID uuid.UUID) (PageTransactionResponse, error)
	CreateTransaction(ctx context.Context, req CreateTransactionRequest, ownerID uuid.UUID) (TransactionResponse, error)
	UpdateTransaction(ctx context.Context, transactionID, ownerID uuid.UUID, updates map[string]any) (TransactionResponse, error)
	DeleteTransaction(ctx context.Context, transactionID, ownerID uuid.UUID) error
}

type Handler struct {
	service transactionService
}

func NewHandler(service transactionService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	txGroup := rg.Group("/transactions")
	{
		txGroup.GET("", h.GetPaginatedTransactions)
		txGroup.POST("", h.CreateTransaction)
		txGroup.PATCH("/:transactionId", h.UpdateTransaction)
		txGroup.DELETE("/:transactionId", h.DeleteTransaction)
	}
}
