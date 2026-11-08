package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// DeleteMyAccount
// @ID           deleteMyAccount
// @Tags         user-controller
// @Security     BearerAuth
// @Success      204  "No Content"
// @Router       /api/v1/users/me [delete]
func (h *Handler) DeleteMyAccount(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), jwtUser); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.Status(nethttp.StatusNoContent)
}
