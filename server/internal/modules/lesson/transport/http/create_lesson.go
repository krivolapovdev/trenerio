package http

import (
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

type StudentVisit = visitmodel.StudentVisit

// @name CreateLessonRequest
type CreateLessonRequest struct {
	GroupID  *uuid.UUID     `json:"groupId" format:"uuid"`
	Date     time.Time      `json:"date" binding:"required" format:"date-time"`
	Students []StudentVisit `json:"students"`
}

// CreateLesson
// @ID           createLesson
// @Tags         lesson-controller
// @Security     BearerAuth
// @Param        body  body      CreateLessonRequest  true  "Create lesson request"
// @Success      201   {object}  LessonResponse
// @Router       /api/v1/lessons [post]
func (h *Handler) CreateLesson(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	var req CreateLessonRequest
	if !web.BindJSON(c, &req) {
		return
	}

	for i := range req.Students {
		req.Students[i].Normalize()
	}

	if err := validator.ValidateDateNotFuture(req.Date, "date"); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	createdLesson, err := h.service.CreateLesson(c.Request.Context(), req.GroupID, req.Date, req.Students, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, LessonResponse(createdLesson))
}
