package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/modules/student/request"
	studentservice "github.com/krivolapovdev/trenerio/internal/modules/student/service"
)

type studentService interface {
	GetAllStudentsSummary(ctx context.Context, ownerID uuid.UUID) ([]studentservice.StudentSummaryResponse, error)
	GetStudentVisits(ctx context.Context, studentID, ownerID uuid.UUID, from, to *time.Time) ([]response.VisitWithLessonResponse, error)
	GetStudentPayments(ctx context.Context, studentID, ownerID uuid.UUID) ([]response.TransactionResponse, error)
	GetStudentDetailsByID(ctx context.Context, studentID, ownerID uuid.UUID) (response.StudentResponse, error)
	CreateStudent(ctx context.Context, req request.CreateStudentRequest, ownerID uuid.UUID) (response.StudentResponse, error)
	CreateStudentPayment(ctx context.Context, studentID, ownerID uuid.UUID, req request.CreateStudentPaymentRequest) (response.TransactionResponse, error)
	UpdateStudent(ctx context.Context, studentID, ownerID uuid.UUID, updates map[string]any) (response.StudentResponse, error)
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
