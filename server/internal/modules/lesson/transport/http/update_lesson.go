package http

import (
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// @name UpdateLessonRequest
type UpdateLessonRequest struct {
	Date     *time.Time     `json:"date" format:"date-time"`
	Students []StudentVisit `json:"students"`
}

// UpdateLesson
// @ID           updateLesson
// @Tags         lesson-controller
// @Security     BearerAuth
// @Param        lessonId  path      string               true  "Lesson ID"  Format(uuid)
// @Param        body      body      UpdateLessonRequest  true  "Update lesson request"
// @Success      200       {object}  LessonResponse
// @Router       /api/v1/lessons/{lessonId} [put]
func (h *Handler) UpdateLesson(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	lessonID, ok := web.ParseUUIDParam(c, "lessonId")
	if !ok {
		return
	}

	var req UpdateLessonRequest
	if !web.BindJSON(c, &req) {
		return
	}

	if req.Date != nil {
		if err := validator.ValidateDateNotFuture(*req.Date, "date"); err != nil {
			pkgerrors.HandleError(c, err)
			return
		}
	}

	updatedLesson, err := h.service.UpdateLesson(c.Request.Context(), lessonID, jwtUser.ID, req.Date, req.Students)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, LessonResponse(updatedLesson))
}
