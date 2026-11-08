package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// UpdateStudent
// @ID           updateStudent
// @Tags         student-controller
// @Security     BearerAuth
// @Param        studentId  path      string             true  "Student ID"        Format(uuid)
// @Param        body       body      map[string]object  true  "Fields to update"
// @Success      200        {object}  StudentResponse
// @Router       /api/v1/students/{studentId} [patch]
func (h *Handler) UpdateStudent(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	studentID, ok := web.ParseUUIDParam(c, "studentId")
	if !ok {
		return
	}

	var updates map[string]any
	if !web.BindJSON(c, &updates) {
		return
	}

	if err := validator.ValidateNameUpdate(updates, "fullName"); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}
	validator.SanitizeTrimmedStringUpdate(updates, "phone")
	validator.SanitizeTrimmedStringUpdate(updates, "note")

	updatedStudent, err := h.service.UpdateStudent(c.Request.Context(), studentID, jwtUser.ID, updates)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	resp := StudentResponse(updatedStudent)
	c.JSON(nethttp.StatusOK, resp)
}
