package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/request"
)

// AddStudents
// @ID           addStudentsToRoster
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        rosterId  path      string                            true  "Roster ID"  Format(uuid)
// @Param        body      body      request.AddRosterStudentsRequest  true  "Students to add"
// @Success      200       {object}  response.RosterResponse
// @Router       /api/v1/rosters/{rosterId}/students [post]
func (h *Handler) AddStudents(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	rosterID, ok := web.ParseUUIDParam(c, "rosterId")
	if !ok {
		return
	}

	var req request.AddRosterStudentsRequest
	if !web.BindJSON(c, &req) {
		return
	}

	updatedRoster, err := h.service.AddStudents(c.Request.Context(), rosterID, jwtUser.ID, req.StudentIds, req.Note)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, updatedRoster)
}
