package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// GetAllGroupsSummary
// @ID           getAllGroupsSummary
// @Tags         group-controller
// @Security     BearerAuth
// @Success      200  {array}   GroupSummaryResponse
// @Router       /api/v1/groups [get]
func (h *Handler) GetAllGroupsSummary(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	summaries, err := h.service.GetAllGroupsSummary(c.Request.Context(), jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, summaries)
}
