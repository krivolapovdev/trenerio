package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// GetRosterByID
// @ID           getRosterById
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        rosterId  path      string  true  "Roster ID"  Format(uuid)
// @Success      200       {object}  response.RosterResponse
// @Router       /api/v1/rosters/{rosterId} [get]
func (h *Handler) GetRosterByID(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	rosterID, ok := web.ParseUUIDParam(c, "rosterId")
	if !ok {
		return
	}

	roster, err := h.service.GetRosterByID(c.Request.Context(), rosterID, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, roster)
}
