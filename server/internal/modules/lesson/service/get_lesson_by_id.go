package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/mapper"
)

func (s *Service) GetLessonByID(ctx context.Context, lessonID, ownerID uuid.UUID) (domain.Lesson, error) {
	logger.InfoContext(ctx, "Getting lesson by id", "lesson_id", lessonID.String(), "user_id", ownerID.String())

	l, err := s.db.FindLessonByIDAndOwnerID(ctx, db.FindLessonByIDAndOwnerIDParams{
		ID:      lessonID,
		OwnerID: ownerID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Lesson not found", "lesson_id", lessonID.String(), "user_id", ownerID.String())
			return domain.Lesson{}, pkgerrors.NewNotFound("Lesson", lessonID, ownerID)
		}

		logger.ErrorContext(ctx, "Failed to get lesson by id", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
		return domain.Lesson{}, err
	}

	logger.DebugContext(ctx, "Retrieved lesson by id", "lesson_id", lessonID.String(), "user_id", ownerID.String())

	return mapper.ToLessonResponse(l), nil
}
