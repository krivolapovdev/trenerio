package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// UpdateGroup
// @ID           updateGroup
// @Tags         group-controller
// @Security     BearerAuth
// @Param        groupId  path      string             true  "Group ID"          Format(uuid)
// @Param        body     body      map[string]object  true  "Fields to update"
// @Success      200      {object}  GroupResponse
// @Router       /api/v1/groups/{groupId} [patch]
func (h *Handler) UpdateGroup(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	groupID, ok := web.ParseUUIDParam(c, "groupId")
	if !ok {
		return
	}

	var updates map[string]any
	if !web.BindJSON(c, &updates) {
		return
	}

	if err := validator.ValidateNameUpdate(updates, "name"); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}
	validator.SanitizeTrimmedStringUpdate(updates, "note")

	updatedGroup, err := h.service.UpdateGroup(c.Request.Context(), groupID, jwtUser.ID, updates)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, updatedGroup)
}
