package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// GetGroupLessons
// @ID           getGroupLessons
// @Tags         group-controller
// @Security     BearerAuth
// @Param        groupId  path      string  true  "Group ID"               Format(uuid)
// @Param        from     query     string  true  "Start date (RFC3339)"   Format(date-time)
// @Param        to       query     string  true  "End date (RFC3339)"     Format(date-time)
// @Success      200      {array}   response.LessonResponse
// @Router       /api/v1/groups/{groupId}/lessons [get]
func (h *Handler) GetGroupLessons(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	groupID, ok := web.ParseUUIDParam(c, "groupId")
	if !ok {
		return
	}

	from, to, ok := web.ParseRequiredRFC3339Range(c, "from", "to", "from and to query parameters are required")
	if !ok {
		return
	}

	lessons, err := h.service.GetGroupLessons(c.Request.Context(), groupID, jwtUser.ID, from, to)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, lessons)
}
