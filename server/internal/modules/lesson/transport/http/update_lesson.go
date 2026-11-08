package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/request"
)

// UpdateLesson
// @ID           updateLesson
// @Tags         lesson-controller
// @Security     BearerAuth
// @Param        lessonId  path      string                      true  "Lesson ID"              Format(uuid)
// @Param        body      body      request.UpdateLessonRequest  true  "Update lesson request"
// @Success      200       {object}  response.LessonResponse
// @Router       /api/v1/lessons/{lessonId} [patch]
func (h *Handler) UpdateLesson(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	lessonID, ok := web.ParseUUIDParam(c, "lessonId")
	if !ok {
		return
	}

	var req request.UpdateLessonRequest
	if !web.BindJSON(c, &req) {
		return
	}

	updatedLesson, err := h.service.UpdateLesson(c.Request.Context(), lessonID, jwtUser.ID, req)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, updatedLesson)
}
