package http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/request"
)

type lessonService interface {
	GetLessonDetailsByID(ctx context.Context, lessonID, ownerID uuid.UUID) (response.LessonDetailsResponse, error)
	CreateLesson(ctx context.Context, req request.CreateLessonRequest, ownerID uuid.UUID) (response.LessonResponse, error)
	UpdateLesson(ctx context.Context, lessonID, ownerID uuid.UUID, req request.UpdateLessonRequest) (response.LessonResponse, error)
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
