package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/db"
	txservice "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

func (s *Service) DeleteLesson(ctx context.Context, lessonID, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Deleting lesson", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		existingLesson, err := qtx.FindLessonByIDAndOwnerID(ctx, db.FindLessonByIDAndOwnerIDParams{
			ID:      lessonID,
			OwnerID: ownerID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Lesson to delete not found", "lesson_id", lessonID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Lesson", lessonID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find lesson for deletion", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		var existingVisits []visitmodel.Visit
		if s.balanceProvider != nil {
			if s.visitProvider != nil {
				existingVisits, err = s.visitProvider.GetVisitsByLessonIDTx(ctx, qtx, lessonID, ownerID)
				if err != nil {
					logger.ErrorContext(ctx, "Failed to fetch visits before lesson deletion", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
					return err
				}
			} else if qtx != nil {
				rows, err := qtx.FindVisitsByLessonIDAndOwnerID(ctx, db.FindVisitsByLessonIDAndOwnerIDParams{
					LessonID: lessonID,
					OwnerID:  ownerID,
				})
				if err != nil {
					return err
				}
				for _, r := range rows {
					existingVisits = append(existingVisits, visitmodel.Visit{
						ID:        r.ID,
						LessonID:  r.LessonID,
						StudentID: r.StudentID,
						Status:    visitmodel.VisitStatus(r.Status),
						Type:      visitmodel.VisitType(r.Type),
						CreatedAt: r.CreatedAt,
					})
				}
			}
		}

		if err := qtx.DeleteLessonByIDAndOwnerID(ctx, db.DeleteLessonByIDAndOwnerIDParams{
			ID:      lessonID,
			OwnerID: ownerID,
		}); err != nil {
			logger.ErrorContext(ctx, "Failed to delete lesson record", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		if len(existingVisits) > 0 && s.balanceProvider != nil {
			studentIDs := make([]uuid.UUID, 0, len(existingVisits))
			for _, ev := range existingVisits {
				studentIDs = append(studentIDs, ev.StudentID)
			}
			studentIDs = utils.UniqueUUIDs(studentIDs)

			freeMap, err := s.fetchStudentFreeFlagsTx(ctx, qtx, studentIDs, ownerID)
			if err != nil {
				return err
			}

			changes := make([]txservice.AttendanceBalanceChange, 0, len(existingVisits))
			for _, ev := range existingVisits {
				changes = append(changes, txservice.AttendanceBalanceChange{
					StudentID:     ev.StudentID,
					OldBillable:   visitmodel.IsBillableVisit(ev.Status, ev.Type, freeMap[ev.StudentID]),
					OldLessonDate: existingLesson.Date,
					NewBillable:   false,
				})
			}

			if err := s.balanceProvider.SyncLessonAttendanceBalancesTx(ctx, qtx, ownerID, changes); err != nil {
				logger.ErrorContext(ctx, "Failed to sync student balances on lesson deletion", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Lesson deleted successfully", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	return nil
}
