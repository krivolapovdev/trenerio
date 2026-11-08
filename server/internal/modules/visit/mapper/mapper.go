package mapper

import (
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

func ToVisit(v db.VisitsModuleVisit) domain.Visit {
	return domain.Visit{
		ID:        v.ID,
		Status:    domain.VisitStatus(v.Status),
		Type:      domain.VisitType(v.Type),
		LessonID:  v.LessonID,
		StudentID: v.StudentID,
		CreatedAt: v.CreatedAt,
	}
}

func ToVisitResponse(v db.VisitsModuleVisit) domain.VisitResponse {
	return ToVisit(v)
}
