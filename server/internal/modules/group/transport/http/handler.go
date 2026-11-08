package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
)

type groupService interface {
	GetAllGroupsSummary(ctx context.Context, ownerID uuid.UUID) ([]domain.GroupSummary, error)
	GetGroupStudents(ctx context.Context, groupID, ownerID uuid.UUID) ([]domain.GroupStudentSummary, error)
	GetGroupLessons(ctx context.Context, groupID, ownerID uuid.UUID, from, to time.Time) ([]lessonmodel.Lesson, error)
	GetGroupReport(ctx context.Context, groupID, ownerID uuid.UUID, year, month int) (domain.GroupReport, error)
	CreateGroup(ctx context.Context, groupToCreate domain.Group, studentIDs []uuid.UUID, ownerID uuid.UUID) (domain.Group, error)
	UpdateGroup(ctx context.Context, groupID, ownerID uuid.UUID, updates map[string]any) (domain.Group, error)
	DeleteGroup(ctx context.Context, groupID, ownerID uuid.UUID) error
}

type Handler struct {
	service groupService
}

func NewHandler(service groupService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	groups := rg.Group("/groups")
	{
		groups.GET("", h.GetAllGroupsSummary)
		groups.GET("/:groupId/students", h.GetGroupStudents)
		groups.GET("/:groupId/lessons", h.GetGroupLessons)
		groups.GET("/:groupId/report", h.GetGroupReport)
		groups.POST("", h.CreateGroup)
		groups.PATCH("/:groupId", h.UpdateGroup)
		groups.DELETE("/:groupId", h.DeleteGroup)
	}
}
