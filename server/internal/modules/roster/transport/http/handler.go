package http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/request"
)

type rosterService interface {
	GetAllRosters(ctx context.Context, ownerID uuid.UUID) ([]response.RosterResponse, error)
	GetRosterByID(ctx context.Context, rosterID, ownerID uuid.UUID) (response.RosterResponse, error)
	GetRosterStudents(ctx context.Context, rosterID, ownerID uuid.UUID) ([]response.StudentResponse, error)
	CreateRoster(ctx context.Context, req request.CreateRosterRequest, ownerID uuid.UUID) (response.RosterResponse, error)
	UpdateRoster(ctx context.Context, rosterID, ownerID uuid.UUID, updates map[string]any) (response.RosterResponse, error)
	DeleteRoster(ctx context.Context, rosterID, ownerID uuid.UUID) error
	AddStudents(ctx context.Context, rosterID, ownerID uuid.UUID, studentIDs []uuid.UUID, note *string) (response.RosterResponse, error)
	UpdateParticipantNote(ctx context.Context, rosterID, studentID, ownerID uuid.UUID, note *string) error
	RemoveStudent(ctx context.Context, rosterID, studentID, ownerID uuid.UUID) error
}

type Handler struct {
	service rosterService
}

func NewHandler(service rosterService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rosters := rg.Group("/rosters")
	{
		rosters.GET("", h.GetAllRosters)
		rosters.POST("", h.CreateRoster)
		rosters.GET("/:rosterId", h.GetRosterByID)
		rosters.PATCH("/:rosterId", h.UpdateRoster)
		rosters.DELETE("/:rosterId", h.DeleteRoster)
		rosters.GET("/:rosterId/students", h.GetRosterStudents)
		rosters.POST("/:rosterId/students", h.AddStudents)
		rosters.PATCH("/:rosterId/students/:studentId", h.UpdateStudentNote)
		rosters.DELETE("/:rosterId/students/:studentId", h.RemoveStudent)
	}
}
