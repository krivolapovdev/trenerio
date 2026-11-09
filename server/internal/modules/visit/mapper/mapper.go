package mapper

import (
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

func ToVisit(v db.VisitsModuleVisit) domain.Visit {
	vType := domain.VisitType(v.Type)
	if domain.VisitStatus(v.Status) == domain.VisitStatusExcused {
		vType = domain.VisitTypeFree
	}
	return domain.Visit{
		ID:        v.ID,
		Status:    domain.VisitStatus(v.Status),
		Type:      vType,
		LessonID:  v.LessonID,
		StudentID: v.StudentID,
		CreatedAt: v.CreatedAt,
	}
}

func ToVisitResponse(v db.VisitsModuleVisit) domain.VisitResponse {
	return ToVisit(v)
}
