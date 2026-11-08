package http

import (
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
)

// @name CreateRosterRequest
type CreateRosterRequest struct {
	Name       string      `json:"name" binding:"required,max=255"`
	Date       *time.Time  `json:"date,omitempty" format:"date-time"`
	Note       *string     `json:"note,omitempty" binding:"omitempty,max=255"`
	StudentIds []uuid.UUID `json:"studentIds,omitempty"`
}

// CreateRoster
// @ID           createRoster
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        body  body      CreateRosterRequest  true  "Create roster request"
// @Success      201   {object}  RosterResponse
// @Router       /api/v1/rosters [post]
func (h *Handler) CreateRoster(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	var req CreateRosterRequest
	if !web.BindJSON(c, &req) {
		return
	}

	trimmedName, err := validator.ValidateRequiredName(req.Name, "name")
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	rosterToCreate := domain.Roster{
		Name: trimmedName,
		Date: req.Date,
		Note: validator.TrimStringPtr(req.Note),
	}

	createdRoster, err := h.service.CreateRoster(c.Request.Context(), rosterToCreate, req.StudentIds, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, createdRoster)
}
