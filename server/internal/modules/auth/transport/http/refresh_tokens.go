package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
)

// @name RefreshTokenRequest
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshTokens
// @ID           refreshTokens
// @Tags         jwt-controller
// @Param        body  body      RefreshTokenRequest  true  "Refresh tokens request"
// @Success      200   {object}  JwtResponse
// @Router       /api/v1/jwt/refresh [post]
func (h *Handler) RefreshTokens(c *gin.Context) {
	var req RefreshTokenRequest
	if !web.BindJSON(c, &req) {
		return
	}

	tokens, err := h.service.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	resp := JwtResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	c.JSON(nethttp.StatusOK, resp)
}
