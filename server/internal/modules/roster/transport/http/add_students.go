package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// @name AddRosterStudentsRequest
type AddRosterStudentsRequest struct {
	StudentIds []uuid.UUID `json:"studentIds" binding:"required"`
	Note       *string     `json:"note,omitempty" binding:"omitempty,max=255"`
}

// AddStudents
// @ID           addStudentsToRoster
// @Tags         roster-controller
// @Security     BearerAuth
// @Param        rosterId  path      string                    true  "Roster ID"  Format(uuid)
// @Param        body      body      AddRosterStudentsRequest  true  "Students to add"
// @Success      200       {object}  RosterResponse
// @Router       /api/v1/rosters/{rosterId}/students [post]
func (h *Handler) AddStudents(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	rosterID, ok := web.ParseUUIDParam(c, "rosterId")
	if !ok {
		return
	}

	var req AddRosterStudentsRequest
	if !web.BindJSON(c, &req) {
		return
	}

	updatedRoster, err := h.service.AddStudents(c.Request.Context(), rosterID, jwtUser.ID, req.StudentIds, req.Note)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, updatedRoster)
}
