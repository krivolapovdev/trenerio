package mapper

import (
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/model"
)

func ToVisitResponse(v db.VisitsModuleVisit) model.VisitResponse {
	return model.VisitResponse{
		ID:        v.ID,
		Status:    model.VisitStatus(v.Status),
		Type:      model.VisitType(v.Type),
		LessonID:  v.LessonID,
		StudentID: v.StudentID,
		CreatedAt: v.CreatedAt,
	}
}
