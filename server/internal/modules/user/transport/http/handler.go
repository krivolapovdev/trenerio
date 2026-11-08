package http

import (
	"context"

	"github.com/gin-gonic/gin"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
)

type userService interface {
	DeleteUser(ctx context.Context, jwtUser authmodel.JwtUser) error
}

type Handler struct {
	service userService
}

func NewHandler(service userService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.DELETE("/me", h.DeleteMyAccount)
	}
}
