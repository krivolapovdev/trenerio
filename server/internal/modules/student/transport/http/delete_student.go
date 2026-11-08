package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// DeleteStudent
// @ID           deleteStudent
// @Tags         student-controller
// @Security     BearerAuth
// @Param        studentId  path  string  true  "Student ID"  Format(uuid)
// @Success      204        "No Content"
// @Router       /api/v1/students/{studentId} [delete]
func (h *Handler) DeleteStudent(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	studentID, ok := web.ParseUUIDParam(c, "studentId")
	if !ok {
		return
	}

	if err := h.service.DeleteStudent(c.Request.Context(), studentID, jwtUser.ID); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.Status(nethttp.StatusNoContent)
}
