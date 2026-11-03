package auth

import "github.com/gin-gonic/gin"

type Router struct {
	handler *Handler
}

func NewRouter(h *Handler) *Router {
	return &Router{handler: h}
}

func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	jwtGroup := rg.Group("/jwt")
	{
		jwtGroup.POST("/refresh", r.handler.RefreshTokens)
	}

	oauth2Group := rg.Group("/oauth2")
	{
		oauth2Group.POST("/google", r.handler.GoogleLogin)
	}

	reviewerGroup := rg.Group("/reviewer")
	{
		reviewerGroup.POST("/login", r.handler.ReviewerLogin)
	}
}
