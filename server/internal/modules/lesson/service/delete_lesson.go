package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func (s *Service) DeleteLesson(ctx context.Context, lessonID, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Deleting lesson", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		if _, err := qtx.FindLessonByIDAndOwnerID(ctx, db.FindLessonByIDAndOwnerIDParams{
			ID:      lessonID,
			OwnerID: ownerID,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Lesson to delete not found", "lesson_id", lessonID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Lesson", lessonID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find lesson for deletion", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		if err := qtx.DeleteLessonByIDAndOwnerID(ctx, db.DeleteLessonByIDAndOwnerIDParams{
			ID:      lessonID,
			OwnerID: ownerID,
		}); err != nil {
			logger.ErrorContext(ctx, "Failed to delete lesson record", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Lesson deleted successfully", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	return nil
}
