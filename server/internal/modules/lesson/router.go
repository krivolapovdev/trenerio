package lesson

import "github.com/gin-gonic/gin"

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	return &Router{handler: handler}
}

// RegisterRoutes registers the lesson routes onto the given router group.
func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	lessons := rg.Group("/lessons")
	{
		lessons.GET("/:lessonId/details", r.handler.GetLessonDetails)
		lessons.POST("", r.handler.CreateLesson)
		lessons.PATCH("/:lessonId", r.handler.UpdateLesson)
		lessons.DELETE("/:lessonId", r.handler.DeleteLesson)
	}
}
