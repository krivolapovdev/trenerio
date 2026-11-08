package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// DeleteGroup
// @ID           deleteGroup
// @Tags         group-controller
// @Security     BearerAuth
// @Param        groupId  path  string  true  "Group ID"  Format(uuid)
// @Success      204      "No Content"
// @Router       /api/v1/groups/{groupId} [delete]
func (h *Handler) DeleteGroup(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	groupID, ok := web.ParseUUIDParam(c, "groupId")
	if !ok {
		return
	}

	if err := h.service.DeleteGroup(c.Request.Context(), groupID, jwtUser.ID); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.Status(nethttp.StatusNoContent)
}
