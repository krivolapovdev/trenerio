package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/harness"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/metric/domain"
	metricsvc "github.com/krivolapovdev/trenerio/internal/modules/metric/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMetricService struct {
	getPaymentStatisticsFn func(ctx context.Context, scopeStr string, startDate, endDate time.Time, ownerID uuid.UUID) ([]domain.PaymentMetric, error)
}

func (m *mockMetricService) GetPaymentStatistics(ctx context.Context, scopeStr string, startDate, endDate time.Time, ownerID uuid.UUID) ([]domain.PaymentMetric, error) {
	if m.getPaymentStatisticsFn != nil {
		return m.getPaymentStatisticsFn(ctx, scopeStr, startDate, endDate, ownerID)
	}
	return nil, nil
}

func setupMetricRouter(svc metricService, user authmodel.JwtUser) *gin.Engine {
	return harness.SetupTestRouter(user, func(rg *gin.RouterGroup) {
		NewHandler(svc).RegisterRoutes(rg)
	})
}

func TestMetricHandler_GetPaymentStatistics(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupMetricRouter(&mockMetricService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/metrics/payments", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("MissingRequiredDates", func(t *testing.T) {
		r := setupMetricRouter(&mockMetricService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/metrics/payments?startDate="+startDate.Format(time.RFC3339), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("InvalidDateFormat", func(t *testing.T) {
		r := setupMetricRouter(&mockMetricService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/metrics/payments?startDate=not-a-date&endDate=not-a-date", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("SuccessWithDefaultScope", func(t *testing.T) {
		expected := []domain.PaymentMetric{
			{Date: startDate, Total: 15000.0},
		}

		mockSvc := &mockMetricService{
			getPaymentStatisticsFn: func(ctx context.Context, scopeStr string, start, end time.Time, ownerID uuid.UUID) ([]domain.PaymentMetric, error) {
				assert.Equal(t, string(metricsvc.Month), scopeStr)
				assert.True(t, startDate.Equal(start))
				assert.True(t, endDate.Equal(end))
				assert.Equal(t, testUser.ID, ownerID)
				return expected, nil
			},
		}

		r := setupMetricRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		url := "/api/v1/metrics/payments?startDate=" + startDate.Format(time.RFC3339) + "&endDate=" + endDate.Format(time.RFC3339)
		req := httptest.NewRequest(nethttp.MethodGet, url, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
		var actual []PaymentMetricResponse
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		require.NoError(t, err)
		assert.Len(t, actual, 1)
		assert.Equal(t, 15000.0, actual[0].Total)
	})

	t.Run("SuccessWithExplicitScope", func(t *testing.T) {
		expected := []domain.PaymentMetric{
			{Date: startDate, Total: 5000.0},
		}

		mockSvc := &mockMetricService{
			getPaymentStatisticsFn: func(ctx context.Context, scopeStr string, start, end time.Time, ownerID uuid.UUID) ([]domain.PaymentMetric, error) {
				assert.Equal(t, string(metricsvc.Week), scopeStr)
				return expected, nil
			},
		}

		r := setupMetricRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		url := "/api/v1/metrics/payments?startDate=" + startDate.Format(time.RFC3339) + "&endDate=" + endDate.Format(time.RFC3339) + "&scope=WEEK"
		req := httptest.NewRequest(nethttp.MethodGet, url, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
		var actual []PaymentMetricResponse
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		require.NoError(t, err)
		assert.Len(t, actual, 1)
		assert.Equal(t, 5000.0, actual[0].Total)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockSvc := &mockMetricService{
			getPaymentStatisticsFn: func(ctx context.Context, scopeStr string, start, end time.Time, ownerID uuid.UUID) ([]domain.PaymentMetric, error) {
				return nil, pkgerrors.NewBadRequest("invalid dates")
			},
		}

		r := setupMetricRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		url := "/api/v1/metrics/payments?startDate=" + startDate.Format(time.RFC3339) + "&endDate=" + endDate.Format(time.RFC3339)
		req := httptest.NewRequest(nethttp.MethodGet, url, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})
}
