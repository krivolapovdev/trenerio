package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// UpdateRoster
// @ID           updateRoster
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        rosterId  path      string             true  "Roster ID"          Format(uuid)
// @Param        body      body      map[string]object  true  "Fields to update"
// @Success      200       {object}  RosterResponse
// @Router       /api/v1/rosters/{rosterId} [patch]
func (h *Handler) UpdateRoster(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	rosterID, ok := web.ParseUUIDParam(c, "rosterId")
	if !ok {
		return
	}

	var updates map[string]any
	if !web.BindJSON(c, &updates) {
		return
	}

	if err := validator.ValidateNameUpdate(updates, "name"); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}
	validator.SanitizeTrimmedStringUpdate(updates, "note")

	updatedRoster, err := h.service.UpdateRoster(c.Request.Context(), rosterID, jwtUser.ID, updates)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, updatedRoster)
}
