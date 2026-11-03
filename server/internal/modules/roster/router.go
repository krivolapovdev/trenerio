package roster

import "github.com/gin-gonic/gin"

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	return &Router{handler: handler}
}

func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	rosters := rg.Group("/rosters")
	{
		rosters.GET("", r.handler.GetAllRosters)
		rosters.GET("/:rosterId", r.handler.GetRosterByID)
		rosters.GET("/:rosterId/students", r.handler.GetRosterStudents)
		rosters.POST("", r.handler.CreateRoster)
		rosters.PATCH("/:rosterId", r.handler.UpdateRoster)
		rosters.DELETE("/:rosterId", r.handler.DeleteRoster)
		rosters.POST("/:rosterId/students", r.handler.AddStudents)
		rosters.PATCH("/:rosterId/students/:studentId", r.handler.UpdateStudentNote)
		rosters.DELETE("/:rosterId/students/:studentId", r.handler.RemoveStudent)
	}
}
