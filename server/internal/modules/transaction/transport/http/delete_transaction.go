package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// DeleteTransaction
// @ID           deleteTransaction
// @Tags         transaction-controller
// @Security     BearerAuth
// @Param        transactionId  path  string  true  "Transaction ID"  Format(uuid)
// @Success      204            "No Content"
// @Router       /api/v1/transactions/{transactionId} [delete]
func (h *Handler) DeleteTransaction(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	txID, ok := web.ParseUUIDParam(c, "transactionId")
	if !ok {
		return
	}

	if err := h.service.DeleteTransaction(c.Request.Context(), txID, jwtUser.ID); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.Status(nethttp.StatusNoContent)
}
