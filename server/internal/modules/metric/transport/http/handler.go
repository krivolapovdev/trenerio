package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/modules/metric/domain"
)

type metricService interface {
	GetPaymentStatistics(ctx context.Context, scopeStr string, startDate, endDate time.Time, ownerID uuid.UUID) ([]domain.PaymentMetric, error)
}

type Handler struct {
	service metricService
}

func NewHandler(service metricService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	metrics := rg.Group("/metrics")
	{
		metrics.GET("/payments", h.GetPaymentStatistics)
	}
}
