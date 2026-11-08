package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
)

// ReviewerLogin
// @ID           login
// @Tags         reviewer-auth-controller
// @Param        X-Reviewer-Key  header    string  true  "Reviewer secret key"
// @Success      200             {object}  response.LoginResponse
// @Router       /api/v1/reviewer/login [post]
func (h *Handler) ReviewerLogin(c *gin.Context) {
	reviewerKey := c.GetHeader(web.HeaderXReviewerKey)
	if reviewerKey == "" {
		pkgerrors.RespondUnauthorized(c, "Authentication failed or token is invalid")
		return
	}

	loginResponse, err := h.service.LoginReviewer(c.Request.Context(), reviewerKey)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, loginResponse)
}
