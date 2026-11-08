package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/mapper"
	txservice "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
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

		oldDate := existingLesson.Date
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
			previousVisits, err := s.visitProvider.UpdateVisitsByLessonIDTx(ctx, qtx, lessonID, uniqueRequests, ownerID)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to update visits for lesson", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
				return err
			}

			if s.balanceProvider != nil {
				oldMap := make(map[uuid.UUID]visitmodel.Visit, len(previousVisits))
				studentIDSet := make([]uuid.UUID, 0, len(previousVisits)+len(uniqueRequests))
				for _, pv := range previousVisits {
					oldMap[pv.StudentID] = pv
					studentIDSet = append(studentIDSet, pv.StudentID)
				}

				newMap := make(map[uuid.UUID]visitmodel.StudentVisit, len(uniqueRequests))
				for _, nv := range uniqueRequests {
					newMap[nv.StudentID] = nv
					studentIDSet = append(studentIDSet, nv.StudentID)
				}
				allStudentIDs := utils.UniqueUUIDs(studentIDSet)

				freeMap, err := s.fetchStudentFreeFlagsTx(ctx, qtx, allStudentIDs, ownerID)
				if err != nil {
					return err
				}

				changes := make([]txservice.AttendanceBalanceChange, 0, len(allStudentIDs))
				for _, sid := range allStudentIDs {
					var oldBillable bool
					if ov, ok := oldMap[sid]; ok {
						oldBillable = visitmodel.IsBillableVisit(ov.Status, ov.Type, freeMap[sid])
					}
					var newBillable bool
					if nv, ok := newMap[sid]; ok {
						newBillable = visitmodel.IsBillableVisit(nv.Status, nv.Type, freeMap[sid])
					}
					changes = append(changes, txservice.AttendanceBalanceChange{
						StudentID:     sid,
						OldBillable:   oldBillable,
						OldLessonDate: oldDate,
						NewBillable:   newBillable,
						NewLessonDate: newDate,
					})
				}

				if err := s.balanceProvider.SyncLessonAttendanceBalancesTx(ctx, qtx, ownerID, changes); err != nil {
					logger.ErrorContext(ctx, "Failed to sync student balances on lesson update", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
					return err
				}
			}
		} else if studentVisits == nil && !newDate.Equal(oldDate) && s.visitProvider != nil && s.balanceProvider != nil {
			existingVisits, err := s.visitProvider.GetVisitsByLessonIDTx(ctx, qtx, lessonID, ownerID)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to fetch visits for lesson date change balance sync", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
			if len(existingVisits) > 0 {
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
					billable := visitmodel.IsBillableVisit(ev.Status, ev.Type, freeMap[ev.StudentID])
					changes = append(changes, txservice.AttendanceBalanceChange{
						StudentID:     ev.StudentID,
						OldBillable:   billable,
						OldLessonDate: oldDate,
						NewBillable:   billable,
						NewLessonDate: newDate,
					})
				}

				if err := s.balanceProvider.SyncLessonAttendanceBalancesTx(ctx, qtx, ownerID, changes); err != nil {
					logger.ErrorContext(ctx, "Failed to sync student balances on lesson date update", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
					return err
				}
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
