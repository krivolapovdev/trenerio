package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/request"
)

// CreateRoster
// @ID           createRoster
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        body  body      request.CreateRosterRequest  true  "Create roster request"
// @Success      201   {object}  response.RosterResponse
// @Router       /api/v1/rosters [post]
func (h *Handler) CreateRoster(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	var req request.CreateRosterRequest
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

	createdRoster, err := h.service.CreateRoster(c.Request.Context(), req, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, createdRoster)
}
