package user

import "github.com/gin-gonic/gin"

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	return &Router{handler: handler}
}

func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.DELETE("/me", r.handler.DeleteMyAccount)
	}
}
