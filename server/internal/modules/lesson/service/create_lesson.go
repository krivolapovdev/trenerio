package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/mapper"
	txservice "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

func (s *Service) CreateLesson(ctx context.Context, groupIDPtr *uuid.UUID, date time.Time, studentVisits []visitmodel.StudentVisit, ownerID uuid.UUID) (domain.Lesson, error) {
	logger.InfoContext(ctx, "Creating lesson", "date", date.Format(time.RFC3339), "user_id", ownerID.String())
	if err := validator.ValidateDateNotFuture(date, "date"); err != nil {
		logger.WarnContext(ctx, "Lesson date cannot be in the future", "date", date.Format(time.RFC3339), "user_id", ownerID.String())
		return domain.Lesson{}, err
	}

	var groupID uuid.NullUUID
	if groupIDPtr != nil {
		groupID = uuid.NullUUID{UUID: *groupIDPtr, Valid: true}
	}

	uniqueVisits := deduplicateStudentVisits(studentVisits)

	var createdLesson db.LessonsModuleLesson
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		var err error
		createdLesson, err = qtx.CreateLesson(ctx, db.CreateLessonParams{
			OwnerID: ownerID,
			GroupID: groupID,
			Date:    date,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create lesson record", "error", err, "user_id", ownerID.String())
			return err
		}

		if len(uniqueVisits) == 0 {
			return nil
		}

		if s.visitProvider != nil {
			if err := s.visitProvider.CreateVisitsTx(ctx, qtx, createdLesson.ID, uniqueVisits, ownerID); err != nil {
				logger.ErrorContext(ctx, "Failed to create visits for lesson", "lesson_id", createdLesson.ID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		if s.balanceProvider != nil {
			studentIDs := make([]uuid.UUID, 0, len(uniqueVisits))
			for _, sv := range uniqueVisits {
				studentIDs = append(studentIDs, sv.StudentID)
			}
			freeMap, err := s.fetchStudentFreeFlagsTx(ctx, qtx, studentIDs, ownerID)
			if err != nil {
				return err
			}

			changes := make([]txservice.AttendanceBalanceChange, 0, len(uniqueVisits))
			for _, sv := range uniqueVisits {
				changes = append(changes, txservice.AttendanceBalanceChange{
					StudentID:     sv.StudentID,
					OldBillable:   false,
					NewBillable:   visitmodel.IsBillableVisit(sv.Status, sv.Type, freeMap[sv.StudentID]),
					NewLessonDate: createdLesson.Date,
				})
			}

			if err := s.balanceProvider.SyncLessonAttendanceBalancesTx(ctx, qtx, ownerID, changes); err != nil {
				logger.ErrorContext(ctx, "Failed to sync student balances on lesson creation", "lesson_id", createdLesson.ID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		return nil
	})
	if err != nil {
		return domain.Lesson{}, err
	}

	logger.InfoContext(ctx, "Lesson created successfully", "lesson_id", createdLesson.ID.String(), "visits_count", len(uniqueVisits), "user_id", ownerID.String())
	return mapper.ToLessonResponse(createdLesson), nil
}
