package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/request"
)

// UpdateStudentNote
// @ID           updateRosterStudentNote
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        rosterId   path  string                                  true  "Roster ID"   Format(uuid)
// @Param        studentId  path  string                                  true  "Student ID"  Format(uuid)
// @Param        body       body  request.UpdateRosterStudentNoteRequest  true  "Student note"
// @Success      204        "No Content"
// @Router       /api/v1/rosters/{rosterId}/students/{studentId} [patch]
func (h *Handler) UpdateStudentNote(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	rosterID, ok := web.ParseUUIDParam(c, "rosterId")
	if !ok {
		return
	}

	studentID, ok := web.ParseUUIDParam(c, "studentId")
	if !ok {
		return
	}

	var req request.UpdateRosterStudentNoteRequest
	if !web.BindJSON(c, &req) {
		return
	}

	if err := h.service.UpdateParticipantNote(c.Request.Context(), rosterID, studentID, jwtUser.ID, req.Note); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.Status(nethttp.StatusNoContent)
}
