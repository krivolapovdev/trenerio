package service

import (
	"github.com/google/uuid"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

func deduplicateStudentVisits(visits []visitmodel.StudentVisit) []visitmodel.StudentVisit {
	if len(visits) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]int, len(visits))
	unique := make([]visitmodel.StudentVisit, 0, len(visits))
	for _, sv := range visits {
		sv.Normalize()
		if idx, exists := seen[sv.StudentID]; exists {
			unique[idx] = sv
		} else {
			seen[sv.StudentID] = len(unique)
			unique = append(unique, sv)
		}
	}
	return unique
}
