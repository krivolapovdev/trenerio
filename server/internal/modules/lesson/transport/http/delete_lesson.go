package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// DeleteLesson
// @ID           deleteLesson
// @Tags         lesson-controller
// @Security     BearerAuth
// @Param        lessonId  path  string  true  "Lesson ID"  Format(uuid)
// @Success      204       "No Content"
// @Router       /api/v1/lessons/{lessonId} [delete]
func (h *Handler) DeleteLesson(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	lessonID, ok := web.ParseUUIDParam(c, "lessonId")
	if !ok {
		return
	}

	if err := h.service.DeleteLesson(c.Request.Context(), lessonID, jwtUser.ID); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.Status(nethttp.StatusNoContent)
}
