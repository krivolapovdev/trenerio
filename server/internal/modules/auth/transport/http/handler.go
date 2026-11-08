package http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/krivolapovdev/trenerio/internal/core/response"
)

type AuthService interface {
	LoginGoogle(ctx context.Context, token string) (response.LoginResponse, error)
	LoginReviewer(ctx context.Context, reviewerKey string) (response.LoginResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (response.JwtResponse, error)
}

type Handler struct {
	service AuthService
}

func NewHandler(service AuthService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	jwtGroup := rg.Group("/jwt")
	{
		jwtGroup.POST("/refresh", h.RefreshTokens)
	}

	oauth2Group := rg.Group("/oauth2")
	{
		oauth2Group.POST("/google", h.GoogleLogin)
	}

	reviewerGroup := rg.Group("/reviewer")
	{
		reviewerGroup.POST("/login", h.ReviewerLogin)
	}
}
