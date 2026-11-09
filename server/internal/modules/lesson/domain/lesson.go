package domain

import (
	"time"

	"github.com/google/uuid"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

// @name LessonResponse
type LessonResponse struct {
	ID        uuid.UUID  `json:"id" binding:"required" format:"uuid"`
	Date      time.Time  `json:"date" binding:"required" format:"date-time"`
	CreatedAt time.Time  `json:"createdAt" binding:"required" format:"date-time"`
	GroupID   *uuid.UUID `json:"groupId,omitempty" format:"uuid"`
}

type Lesson = LessonResponse

type LessonDetails struct {
	ID            uuid.UUID              `json:"id" binding:"required" format:"uuid"`
	Date          time.Time              `json:"date" binding:"required" format:"date-time"`
	CreatedAt     time.Time              `json:"createdAt" binding:"required" format:"date-time"`
	GroupID       *uuid.UUID             `json:"groupId,omitempty" format:"uuid"`
	StudentVisits []visitmodel.Visit     `json:"studentVisits" binding:"required"`
	Students      []studentmodel.Student `json:"students" binding:"required"`
}
