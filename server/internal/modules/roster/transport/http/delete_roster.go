package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// DeleteRoster
// @ID           deleteRoster
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        rosterId  path  string  true  "Roster ID"  Format(uuid)
// @Success      204       "No Content"
// @Router       /api/v1/rosters/{rosterId} [delete]
func (h *Handler) DeleteRoster(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	rosterID, ok := web.ParseUUIDParam(c, "rosterId")
	if !ok {
		return
	}

	if err := h.service.DeleteRoster(c.Request.Context(), rosterID, jwtUser.ID); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.Status(nethttp.StatusNoContent)
}
