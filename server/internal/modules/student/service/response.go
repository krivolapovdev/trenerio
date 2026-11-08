package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/core/types"
	"github.com/krivolapovdev/trenerio/internal/modules/student/model"
)

// @name StudentSummaryResponse
type StudentSummaryResponse struct {
	ID            uuid.UUID                `json:"id" binding:"required" format:"uuid"`
	FullName      string                   `json:"fullName" binding:"required"`
	Birthdate     *types.Date              `json:"birthdate,omitempty" swaggertype:"string" format:"date"`
	Phone         *string                  `json:"phone,omitempty"`
	Note          *string                  `json:"note,omitempty"`
	CreatedAt     time.Time                `json:"createdAt" binding:"required" format:"date-time"`
	Free          bool                     `json:"free" binding:"required"`
	Statuses      []model.StudentStatus    `json:"statuses" binding:"required"`
	StudentGroups []response.GroupResponse `json:"studentGroups"`
	PaidUntil     *time.Time               `json:"paidUntil,omitempty" format:"date-time"`
	SessionsLeft  int64                    `json:"sessionsLeft" default:"0"`
}
