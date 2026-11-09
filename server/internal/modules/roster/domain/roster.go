package domain

import (
	"time"

	"github.com/google/uuid"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

// @name Roster
type Roster struct {
	ID              uuid.UUID              `json:"id" binding:"required" format:"uuid"`
	Name            string                 `json:"name" binding:"required"`
	Date            *time.Time             `json:"date,omitempty" format:"date-time"`
	Note            *string                `json:"note,omitempty"`
	CreatedAt       time.Time              `json:"createdAt" binding:"required" format:"date-time"`
	CountOfStudents int64                  `json:"countOfStudents"`
	Students        []studentmodel.Student `json:"students,omitempty"`
}

type Participant struct {
	ID        uuid.UUID `json:"id" format:"uuid"`
	RosterID  uuid.UUID `json:"rosterId" format:"uuid"`
	StudentID uuid.UUID `json:"studentId" format:"uuid"`
	Note      *string   `json:"note,omitempty"`
}
