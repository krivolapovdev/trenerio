package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	"github.com/krivolapovdev/trenerio/internal/modules/auth/request"
)

// GoogleLogin
// @ID           googleLogin
// @Tags         o-auth-2-controller
// @Param        body  body      request.OAuth2LoginRequest  true  "Google OAuth2 login request"
// @Success      200   {object}  response.LoginResponse
// @Router       /api/v1/oauth2/google [post]
func (h *Handler) GoogleLogin(c *gin.Context) {
	var req request.OAuth2LoginRequest
	if !web.BindJSON(c, &req) {
		return
	}

	loginResponse, err := h.service.LoginGoogle(c.Request.Context(), req.Token)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, loginResponse)
}
