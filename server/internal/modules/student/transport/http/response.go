package http

import (
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/types"
	groupmodel "github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

// @name StudentResponse
type StudentResponse = domain.Student

// @name StudentSummaryResponse
type StudentSummaryResponse struct {
	ID            uuid.UUID              `json:"id" binding:"required" format:"uuid"`
	FullName      string                 `json:"fullName" binding:"required"`
	Birthdate     *types.Date            `json:"birthdate,omitempty" swaggertype:"string" format:"date"`
	Phone         *string                `json:"phone,omitempty"`
	Note          *string                `json:"note,omitempty"`
	CreatedAt     time.Time              `json:"createdAt" binding:"required" format:"date-time"`
	Free          bool                   `json:"free" binding:"required"`
	Statuses      []domain.StudentStatus `json:"statuses" binding:"required"`
	StudentGroups []groupmodel.Group     `json:"studentGroups"`
	PaidUntil     *time.Time             `json:"paidUntil,omitempty" format:"date-time"`
	SessionsLeft  int64                  `json:"sessionsLeft" default:"0"`
}

// @name VisitWithLessonResponse
type VisitWithLessonResponse struct {
	Visit  visitmodel.Visit   `json:"visit" binding:"required"`
	Lesson lessonmodel.Lesson `json:"lesson" binding:"required"`
}
