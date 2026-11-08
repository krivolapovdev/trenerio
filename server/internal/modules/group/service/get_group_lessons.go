package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
)

func (s *Service) GetGroupLessons(ctx context.Context, groupID, ownerID uuid.UUID, from, to time.Time) ([]lessonmodel.Lesson, error) {
	logger.InfoContext(ctx, "Getting lessons for group", "group_id", groupID.String(), "user_id", ownerID.String())
	if _, err := s.GetGroupByID(ctx, groupID, ownerID); err != nil {
		logger.WarnContext(ctx, "Group not found for getting lessons", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	var lessons []lessonmodel.Lesson
	if s.lessonProvider != nil {
		var err error
		lessons, err = s.lessonProvider.GetLessonsByGroupIDAndDateRange(ctx, groupID, ownerID, from, to)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch lessons for group", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
			return nil, err
		}
	}
	logger.DebugContext(ctx, "Retrieved group lessons", "group_id", groupID.String(), "count", len(lessons), "user_id", ownerID.String())
	return lessons, nil
}
