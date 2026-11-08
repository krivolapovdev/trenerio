package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
)

// @name OAuth2LoginRequest
type OAuth2LoginRequest struct {
	Token string `json:"token" binding:"required"`
}

// GoogleLogin
// @ID           googleLogin
// @Tags         o-auth-2-controller
// @Param        body  body      OAuth2LoginRequest  true  "Google OAuth2 login request"
// @Success      200   {object}  LoginResponse
// @Router       /api/v1/oauth2/google [post]
func (h *Handler) GoogleLogin(c *gin.Context) {
	var req OAuth2LoginRequest
	if !web.BindJSON(c, &req) {
		return
	}

	result, err := h.service.LoginGoogle(c.Request.Context(), req.Token)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	resp := LoginResponse{
		User: UserResponse{
			ID:    result.User.ID,
			Email: result.User.Email,
		},
		JwtTokens: JwtResponse{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
		},
	}

	c.JSON(nethttp.StatusOK, resp)
}
