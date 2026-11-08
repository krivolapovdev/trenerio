package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/group/request"
)

// CreateGroup
// @ID           createGroup
// @Tags         group-controller
// @Security     BearerAuth
// @Param        body  body      request.CreateGroupRequest  true  "Create group request"
// @Success      201   {object}  response.GroupResponse
// @Router       /api/v1/groups [post]
func (h *Handler) CreateGroup(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	var req request.CreateGroupRequest
	if !web.BindJSON(c, &req) {
		return
	}

	trimmedName, err := validator.ValidateRequiredName(req.Name, "name")
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}
	req.Name = trimmedName
	req.Note = validator.TrimStringPtr(req.Note)

	createdGroup, err := h.service.CreateGroup(c.Request.Context(), req, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, createdGroup)
}
