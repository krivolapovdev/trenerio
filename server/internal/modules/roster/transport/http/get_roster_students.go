package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// GetRosterStudents
// @ID           getRosterStudents
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        rosterId  path      string  true  "Roster ID"  Format(uuid)
// @Success      200       {array}   response.StudentResponse
// @Router       /api/v1/rosters/{rosterId}/students [get]
func (h *Handler) GetRosterStudents(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	rosterID, ok := web.ParseUUIDParam(c, "rosterId")
	if !ok {
		return
	}

	students, err := h.service.GetRosterStudents(c.Request.Context(), rosterID, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, students)
}
