package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// GetLessonDetails
// @ID           getLessonDetails
// @Tags         lesson-controller
// @Security     BearerAuth
// @Param        lessonId  path      string  true  "Lesson ID"  Format(uuid)
// @Success      200       {object}  LessonDetailsResponse
// @Router       /api/v1/lessons/{lessonId}/details [get]
func (h *Handler) GetLessonDetails(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	lessonID, ok := web.ParseUUIDParam(c, "lessonId")
	if !ok {
		return
	}

	details, err := h.service.GetLessonDetailsByID(c.Request.Context(), lessonID, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, details)
}
