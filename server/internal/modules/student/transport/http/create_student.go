package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/student/request"
)

// CreateStudent
// @ID           createStudent
// @Tags         student-controller
// @Security     BearerAuth
// @Param        body  body      request.CreateStudentRequest  true  "Create student request"
// @Success      201   {object}  response.StudentResponse
// @Router       /api/v1/students [post]
func (h *Handler) CreateStudent(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	var req request.CreateStudentRequest
	if !web.BindJSON(c, &req) {
		return
	}

	var err error
	req.FullName, err = validator.ValidateRequiredName(req.FullName, "fullName")
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}
	req.Phone = validator.TrimStringPtr(req.Phone)
	req.Note = validator.TrimStringPtr(req.Note)

	createdStudent, err := h.service.CreateStudent(c.Request.Context(), req, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, createdStudent)
}
