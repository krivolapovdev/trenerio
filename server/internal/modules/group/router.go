package group

import "github.com/gin-gonic/gin"

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	return &Router{handler: handler}
}

func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	groups := rg.Group("/groups")
	{
		groups.GET("", r.handler.GetAllGroupsSummary)
		groups.GET("/:groupId/students", r.handler.GetGroupStudents)
		groups.GET("/:groupId/lessons", r.handler.GetGroupLessons)
		groups.GET("/:groupId/report", r.handler.GetGroupReport)
		groups.POST("", r.handler.CreateGroup)
		groups.PATCH("/:groupId", r.handler.UpdateGroup)
		groups.DELETE("/:groupId", r.handler.DeleteGroup)
	}
}
