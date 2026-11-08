package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/request"
)

// CreateLesson
// @ID           createLesson
// @Tags         lesson-controller
// @Security     BearerAuth
// @Param        body  body      request.CreateLessonRequest  true  "Create lesson request"
// @Success      201   {object}  response.LessonResponse
// @Router       /api/v1/lessons [post]
func (h *Handler) CreateLesson(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	var req request.CreateLessonRequest
	if !web.BindJSON(c, &req) {
		return
	}

	createdLesson, err := h.service.CreateLesson(c.Request.Context(), req, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, createdLesson)
}
