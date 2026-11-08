package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	"github.com/krivolapovdev/trenerio/internal/modules/auth/request"
)

// RefreshTokens
// @ID           refreshTokens
// @Tags         jwt-controller
// @Param        body  body      request.RefreshTokenRequest  true  "Refresh token request"
// @Success      200   {object}  response.JwtResponse
// @Router       /api/v1/jwt/refresh [post]
func (h *Handler) RefreshTokens(c *gin.Context) {
	var req request.RefreshTokenRequest
	if !web.BindJSON(c, &req) {
		return
	}

	jwtTokens, err := h.service.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, jwtTokens)
}
