package domain

import "github.com/google/uuid"

// @name StudentVisit
type StudentVisit struct {
	StudentID uuid.UUID   `json:"studentId" format:"uuid"`
	Status    VisitStatus `json:"status"`
	Type      VisitType   `json:"type"`
}

// Normalize ensures invariants for StudentVisit:
// If visit status is EXCUSED, visit type is always FREE.
func (sv *StudentVisit) Normalize() {
	if sv.Status == VisitStatusExcused {
		sv.Type = VisitTypeFree
	}
}
