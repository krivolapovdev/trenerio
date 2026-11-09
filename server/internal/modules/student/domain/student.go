package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/types"
)

// @name StudentStatus
type StudentStatus string

const (
	StudentStatusInactive StudentStatus = "INACTIVE"
	StudentStatusPresent  StudentStatus = "PRESENT"
	StudentStatusMissing  StudentStatus = "MISSING"
	StudentStatusPaid     StudentStatus = "PAID"
	StudentStatusUnpaid   StudentStatus = "UNPAID"
	StudentStatusFree     StudentStatus = "FREE"
)

// @name Student
type Student struct {
	ID        uuid.UUID   `json:"id" binding:"required" format:"uuid"`
	FullName  string      `json:"fullName" binding:"required"`
	Birthdate *types.Date `json:"birthdate,omitempty" format:"date"`
	Phone     *string     `json:"phone,omitempty"`
	Note      *string     `json:"note,omitempty"`
	CreatedAt time.Time   `json:"createdAt" binding:"required" format:"date-time"`
	Free      bool        `json:"free" binding:"required"`
}
