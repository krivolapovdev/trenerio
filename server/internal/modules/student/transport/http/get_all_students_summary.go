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
// @Success      200  {array}   service.StudentSummaryResponse
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

	c.JSON(nethttp.StatusOK, summaries)
}
