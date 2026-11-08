package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func (s *Service) RemoveVisitsByLessonID(ctx context.Context, lessonID, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Removing visits by lesson id", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	err := s.db.DeleteVisitsByLessonIDAndOwnerID(ctx, db.DeleteVisitsByLessonIDAndOwnerIDParams{
		LessonID: lessonID,
		OwnerID:  ownerID,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to remove visits by lesson id", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
		return err
	}
	logger.InfoContext(ctx, "Visits removed successfully", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	return nil
}
