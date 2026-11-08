package mapper

import (
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
)

func ToLessonModel(l db.LessonsModuleLesson) domain.Lesson {
	var groupID *uuid.UUID
	if l.GroupID.Valid {
		groupID = &l.GroupID.UUID
	}
	return domain.Lesson{
		ID:        l.ID,
		Date:      l.Date,
		CreatedAt: l.CreatedAt,
		GroupID:   groupID,
	}
}

func ToLessonResponse(l db.LessonsModuleLesson) domain.Lesson {
	return ToLessonModel(l)
}
