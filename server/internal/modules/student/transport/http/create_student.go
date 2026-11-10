package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/types"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

// @name CreateStudentRequest
type CreateStudentRequest struct {
	FullName  string      `json:"fullName" binding:"required,max=255"`
	Birthdate *types.Date `json:"birthdate,omitempty" swaggertype:"string" format:"date"`
	Phone     *string     `json:"phone,omitempty" binding:"omitempty,max=255"`
	Note      *string     `json:"note,omitempty" binding:"omitempty,max=255"`
	Free      bool        `json:"free"`
}

// CreateStudent
// @ID           createStudent
// @Tags         student-controller
// @Security     BearerAuth
// @Param        body  body      CreateStudentRequest  true  "Create student request"
// @Success      201   {object}  StudentResponse
// @Router       /api/v1/students [post]
func (h *Handler) CreateStudent(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	var req CreateStudentRequest
	if !web.BindJSON(c, &req) {
		return
	}

	fullName, err := validator.ValidateRequiredName(req.FullName, "fullName")
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	studentToCreate := domain.Student{
		FullName:  fullName,
		Birthdate: req.Birthdate,
		Phone:     validator.TrimStringPtr(req.Phone),
		Note:      validator.TrimStringPtr(req.Note),
		Free:      req.Free,
	}

	createdStudent, err := h.service.CreateStudent(c.Request.Context(), studentToCreate, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	resp := &StudentResponse{
		ID:        createdStudent.ID,
		FullName:  createdStudent.FullName,
		Birthdate: createdStudent.Birthdate,
		Phone:     createdStudent.Phone,
		Note:      createdStudent.Note,
		CreatedAt: createdStudent.CreatedAt,
		Free:      createdStudent.Free,
	}

	c.JSON(nethttp.StatusCreated, resp)
}
