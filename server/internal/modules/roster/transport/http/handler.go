package http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

type rosterService interface {
	GetAllRosters(ctx context.Context, ownerID uuid.UUID) ([]domain.Roster, error)
	GetRosterByID(ctx context.Context, rosterID, ownerID uuid.UUID) (domain.Roster, error)
	GetRosterStudents(ctx context.Context, rosterID, ownerID uuid.UUID) ([]studentmodel.Student, error)
	CreateRoster(ctx context.Context, rosterToCreate domain.Roster, studentIDs []uuid.UUID, ownerID uuid.UUID) (domain.Roster, error)
	UpdateRoster(ctx context.Context, rosterID, ownerID uuid.UUID, updates map[string]any) (domain.Roster, error)
	DeleteRoster(ctx context.Context, rosterID, ownerID uuid.UUID) error
	AddStudents(ctx context.Context, rosterID, ownerID uuid.UUID, studentIDs []uuid.UUID, note *string) (domain.Roster, error)
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
