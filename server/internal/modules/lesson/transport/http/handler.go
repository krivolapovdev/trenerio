package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

type lessonService interface {
	GetLessonDetailsByID(ctx context.Context, lessonID, ownerID uuid.UUID) (domain.LessonDetails, error)
	CreateLesson(ctx context.Context, groupID *uuid.UUID, date time.Time, studentVisits []visitmodel.StudentVisit, ownerID uuid.UUID) (domain.Lesson, error)
	UpdateLesson(ctx context.Context, lessonID, ownerID uuid.UUID, date *time.Time, studentVisits []visitmodel.StudentVisit) (domain.Lesson, error)
	DeleteLesson(ctx context.Context, lessonID, ownerID uuid.UUID) error
}

type Handler struct {
	service lessonService
}

func NewHandler(service lessonService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	lessons := rg.Group("/lessons")
	{
		lessons.GET("/:lessonId/details", h.GetLessonDetails)
		lessons.POST("", h.CreateLesson)
		lessons.PATCH("/:lessonId", h.UpdateLesson)
		lessons.DELETE("/:lessonId", h.DeleteLesson)
	}
}
