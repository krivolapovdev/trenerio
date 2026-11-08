package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// GetPaginatedTransactions
// @ID           getPaginatedTransactions
// @Tags         transaction-controller
// @Security     BearerAuth
// @Param        page  query     int  false  "Page number (1-indexed)"  default(1)
// @Param        size  query     int  false  "Page size"                default(20)
// @Success      200   {object}  response.PageTransactionResponse
// @Router       /api/v1/transactions [get]
func (h *Handler) GetPaginatedTransactions(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	page, size := web.ParsePagination(c, 1, 20)

	pageResponse, err := h.service.GetPaginatedTransactions(c.Request.Context(), page, size, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, pageResponse)
}
