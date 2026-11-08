package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// GetAllStudentsSummary
// @ID           getStudentsSummary
// @Tags         student-controller
// @Security     BearerAuth
// @Success      200  {array}   StudentSummaryResponse
// @Router       /api/v1/students/overview [get]
func (h *Handler) GetAllStudentsSummary(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	summaries, err := h.service.GetAllStudentsSummary(c.Request.Context(), jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	resp := make([]StudentSummaryResponse, len(summaries))
	for i, s := range summaries {
		resp[i] = StudentSummaryResponse{
			ID:            s.ID,
			FullName:      s.FullName,
			Birthdate:     s.Birthdate,
			Phone:         s.Phone,
			Note:          s.Note,
			CreatedAt:     s.CreatedAt,
			Free:          s.Free,
			Statuses:      s.Statuses,
			StudentGroups: s.StudentGroups,
			PaidUntil:     s.PaidUntil,
			SessionsLeft:  s.SessionsLeft,
		}
	}

	c.JSON(nethttp.StatusOK, resp)
}
