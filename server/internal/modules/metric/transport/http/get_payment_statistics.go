package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	metricservice "github.com/krivolapovdev/trenerio/internal/modules/metric/service"
)

// GetPaymentStatistics
// @ID           getPaymentStatistics
// @Tags         metric-controller
// @Security     BearerAuth
// @Param        scope      query     string  false  "Aggregation scope"      Enums(WEEK, MONTH, YEAR)  default(MONTH)
// @Param        startDate  query     string  true   "Start date (RFC3339)"   Format(date-time)
// @Param        endDate    query     string  true   "End date (RFC3339)"     Format(date-time)
// @Success      200        {array}   PaymentMetricResponse
// @Router       /api/v1/metrics/payments [get]
func (h *Handler) GetPaymentStatistics(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	scope := c.DefaultQuery("scope", string(metricservice.Month))
	startDate, endDate, ok := web.ParseRequiredRFC3339Range(c, "startDate", "endDate", "startDate and endDate are required")
	if !ok {
		return
	}

	stats, err := h.service.GetPaymentStatistics(c.Request.Context(), scope, startDate, endDate, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, stats)
}
