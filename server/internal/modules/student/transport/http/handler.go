package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	studentservice "github.com/krivolapovdev/trenerio/internal/modules/student/service"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
)

type studentService interface {
	GetAllStudentsSummary(ctx context.Context, ownerID uuid.UUID) ([]studentservice.StudentSummary, error)
	GetStudentVisits(ctx context.Context, studentID, ownerID uuid.UUID, from, to *time.Time) ([]studentservice.StudentVisitWithLesson, error)
	GetStudentPayments(ctx context.Context, studentID, ownerID uuid.UUID) ([]txmodel.Transaction, error)
	GetStudentDetailsByID(ctx context.Context, studentID, ownerID uuid.UUID) (studentmodel.Student, error)
	CreateStudent(ctx context.Context, studentToCreate studentmodel.Student, ownerID uuid.UUID) (studentmodel.Student, error)
	CreateStudentPayment(ctx context.Context, studentID, ownerID uuid.UUID, amount float64, date time.Time, paidFrom, paidUntil *time.Time, sessionsAdded *int64) (txmodel.Transaction, error)
	UpdateStudent(ctx context.Context, studentID, ownerID uuid.UUID, updates map[string]any) (studentmodel.Student, error)
	DeleteStudent(ctx context.Context, studentID, ownerID uuid.UUID) error
}

type Handler struct {
	service studentService
}

func NewHandler(service studentService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	students := rg.Group("/students")
	{
		students.GET("", h.GetAllStudentsSummary)
		students.GET("/overview", h.GetAllStudentsSummary)
		students.GET("/:studentId/visits", h.GetStudentVisits)
		students.GET("/:studentId/payments", h.GetStudentPayments)
		students.POST("", h.CreateStudent)
		students.POST("/:studentId/payments", h.CreateStudentPayment)
		students.PATCH("/:studentId", h.UpdateStudent)
		students.DELETE("/:studentId", h.DeleteStudent)
	}
}
