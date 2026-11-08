package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/mapper"
)

func (s *Service) GetLessonsByIDs(ctx context.Context, lessonIDs []uuid.UUID, ownerID uuid.UUID) ([]domain.Lesson, error) {
	if len(lessonIDs) == 0 {
		logger.DebugContext(ctx, "No lesson IDs provided to get lessons", "user_id", ownerID.String())
		return []domain.Lesson{}, nil
	}
	rows, err := s.db.FindLessonsByIDsAndOwnerID(ctx, db.FindLessonsByIDsAndOwnerIDParams{
		OwnerID: ownerID,
		Column2: lessonIDs,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get lessons by IDs", "count", len(lessonIDs), "error", err, "user_id", ownerID.String())
		return nil, err
	}
	lessons := make([]domain.Lesson, 0, len(rows))
	for _, r := range rows {
		lessons = append(lessons, mapper.ToLessonResponse(r))
	}
	logger.DebugContext(ctx, "Retrieved lessons by IDs", "count", len(lessons), "user_id", ownerID.String())
	return lessons, nil
}
