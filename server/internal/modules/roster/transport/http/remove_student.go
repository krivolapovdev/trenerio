package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// RemoveStudent
// @ID           removeStudentFromRoster
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        rosterId   path  string  true  "Roster ID"   Format(uuid)
// @Param        studentId  path  string  true  "Student ID"  Format(uuid)
// @Success      204        "No Content"
// @Router       /api/v1/rosters/{rosterId}/students/{studentId} [delete]
func (h *Handler) RemoveStudent(c *gin.Context) {
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

	if err := h.service.RemoveStudent(c.Request.Context(), rosterID, studentID, jwtUser.ID); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.Status(nethttp.StatusNoContent)
}
