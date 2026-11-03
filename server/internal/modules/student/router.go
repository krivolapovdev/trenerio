package student

import "github.com/gin-gonic/gin"

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	return &Router{handler: handler}
}

func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	students := rg.Group("/students")
	{
		students.GET("", r.handler.GetAllStudentsSummary)
		students.GET("/overview", r.handler.GetAllStudentsSummary)
		students.GET("/:studentId/visits", r.handler.GetStudentVisits)
		students.GET("/:studentId/payments", r.handler.GetStudentPayments)
		students.POST("", r.handler.CreateStudent)
		students.POST("/:studentId/payments", r.handler.CreateStudentPayment)
		students.PATCH("/:studentId", r.handler.UpdateStudent)
		students.DELETE("/:studentId", r.handler.DeleteStudent)
	}
}
