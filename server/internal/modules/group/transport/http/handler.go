package http

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/modules/group/request"
	groupservice "github.com/krivolapovdev/trenerio/internal/modules/group/service"
)

type groupService interface {
	GetAllGroupsSummary(ctx context.Context, ownerID uuid.UUID) ([]response.GroupSummaryResponse, error)
	GetGroupStudents(ctx context.Context, groupID, ownerID uuid.UUID) ([]groupservice.GroupStudentSummaryResponse, error)
	GetGroupLessons(ctx context.Context, groupID, ownerID uuid.UUID, from, to time.Time) ([]response.LessonResponse, error)
	GetGroupReport(ctx context.Context, groupID, ownerID uuid.UUID, year, month int) (groupservice.GroupReportResponse, error)
	CreateGroup(ctx context.Context, req request.CreateGroupRequest, ownerID uuid.UUID) (response.GroupResponse, error)
	UpdateGroup(ctx context.Context, groupID, ownerID uuid.UUID, updates map[string]any) (response.GroupResponse, error)
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
