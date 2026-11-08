package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/mapper"
)

func (s *Service) GetLessonsByGroupIDAndDateRange(ctx context.Context, groupID, ownerID uuid.UUID, from, to time.Time) ([]domain.Lesson, error) {
	logger.InfoContext(ctx, "Getting lessons by group id and date range", "group_id", groupID.String(), "user_id", ownerID.String())
	rows, err := s.db.FindLessonsByGroupIDAndDateBetween(ctx, db.FindLessonsByGroupIDAndDateBetweenParams{
		GroupID: uuid.NullUUID{UUID: groupID, Valid: true},
		OwnerID: ownerID,
		Date:    from,
		Date_2:  to,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get lessons by group id and date range", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}
	lessons := make([]domain.Lesson, 0, len(rows))
	for _, r := range rows {
		lessons = append(lessons, mapper.ToLessonResponse(r))
	}
	logger.DebugContext(ctx, "Retrieved lessons by group id and date range", "group_id", groupID.String(), "count", len(lessons), "user_id", ownerID.String())
	return lessons, nil
}
