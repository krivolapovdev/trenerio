package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// GetAllRosters
// @ID           getAllRosters
// @Tags         roster-controller
// @Security     BearerAuth
// @Success      200  {array}   RosterResponse
// @Router       /api/v1/rosters [get]
func (h *Handler) GetAllRosters(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	rosters, err := h.service.GetAllRosters(c.Request.Context(), jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, rosters)
}
