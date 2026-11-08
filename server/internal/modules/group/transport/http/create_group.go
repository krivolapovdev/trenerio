package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
)

// @name CreateGroupRequest
type CreateGroupRequest struct {
	Name         string      `json:"name" binding:"required,max=255"`
	DefaultPrice *float64    `json:"defaultPrice" binding:"omitempty,gt=0"`
	Note         *string     `json:"note" binding:"omitempty,max=255"`
	StudentIds   []uuid.UUID `json:"studentIds"`
}

// CreateGroup
// @ID           createGroup
// @Tags         group-controller
// @Security     BearerAuth
// @Param        body  body      CreateGroupRequest  true  "Create group request"
// @Success      201   {object}  GroupResponse
// @Router       /api/v1/groups [post]
func (h *Handler) CreateGroup(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	var req CreateGroupRequest
	if !web.BindJSON(c, &req) {
		return
	}

	trimmedName, err := validator.ValidateRequiredName(req.Name, "name")
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	groupToCreate := domain.Group{
		Name:         trimmedName,
		DefaultPrice: req.DefaultPrice,
		Note:         validator.TrimStringPtr(req.Note),
	}

	createdGroup, err := h.service.CreateGroup(c.Request.Context(), groupToCreate, req.StudentIds, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, createdGroup)
}
