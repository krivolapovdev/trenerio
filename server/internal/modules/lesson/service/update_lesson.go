package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/mapper"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

func (s *Service) UpdateLesson(ctx context.Context, lessonID, ownerID uuid.UUID, datePtr *time.Time, studentVisits []visitmodel.StudentVisit) (domain.Lesson, error) {
	logger.InfoContext(ctx, "Updating lesson", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	if datePtr != nil {
		if err := validator.ValidateDateNotFuture(*datePtr, "date"); err != nil {
			logger.WarnContext(ctx, "Updated lesson date cannot be in the future", "date", datePtr.Format(time.RFC3339), "user_id", ownerID.String())
			return domain.Lesson{}, err
		}
	}

	var updatedLesson db.LessonsModuleLesson
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		existingLesson, err := qtx.FindLessonByIDAndOwnerID(ctx, db.FindLessonByIDAndOwnerIDParams{
			ID:      lessonID,
			OwnerID: ownerID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Lesson to update not found", "lesson_id", lessonID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Lesson", lessonID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find lesson for update", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		newDate := existingLesson.Date
		if datePtr != nil {
			newDate = *datePtr
		}

		updatedLesson, err = qtx.UpdateLesson(ctx, db.UpdateLessonParams{
			ID:      lessonID,
			OwnerID: ownerID,
			Date:    newDate,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to update lesson record", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		if studentVisits != nil && s.visitProvider != nil {
			uniqueRequests := deduplicateStudentVisits(studentVisits)
			if _, err := s.visitProvider.UpdateVisitsByLessonIDTx(ctx, qtx, lessonID, uniqueRequests, ownerID); err != nil {
				logger.ErrorContext(ctx, "Failed to update visits for lesson", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		return nil
	})
	if err != nil {
		return domain.Lesson{}, err
	}

	logger.InfoContext(ctx, "Lesson updated successfully", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	return mapper.ToLessonResponse(updatedLesson), nil
}
