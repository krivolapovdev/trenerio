package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/mapper"
)

func (s *Service) UpdateVisitsByLessonID(ctx context.Context, lessonID uuid.UUID, requests []StudentVisit, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Updating visits by lesson id", "lesson_id", lessonID.String(), "count", len(requests), "user_id", ownerID.String())
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		_, err := s.UpdateVisitsByLessonIDTx(ctx, qtx, lessonID, requests, ownerID)
		return err
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update visits in transaction", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
		return err
	}
	logger.InfoContext(ctx, "Visits updated successfully", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	return nil
}

func (s *Service) UpdateVisitsByLessonIDTx(ctx context.Context, qtx *db.Queries, lessonID uuid.UUID, requests []StudentVisit, ownerID uuid.UUID) ([]VisitResponse, error) {
	existingVisits, err := qtx.FindVisitsByLessonIDAndOwnerID(ctx, db.FindVisitsByLessonIDAndOwnerIDParams{
		LessonID: lessonID,
		OwnerID:  ownerID,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch existing visits for lesson update", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	previousResponses := make([]VisitResponse, 0, len(existingVisits))
	existingMap := make(map[uuid.UUID]db.VisitsModuleVisit, len(existingVisits))
	for _, ev := range existingVisits {
		existingMap[ev.StudentID] = ev
		previousResponses = append(previousResponses, mapper.ToVisitResponse(ev))
	}

	reqMap := make(map[uuid.UUID]StudentVisit, len(requests))
	for _, req := range requests {
		reqMap[req.StudentID] = req
	}

	deletedCount := 0
	for studentID, ev := range existingMap {
		if _, ok := reqMap[studentID]; !ok {
			if err := qtx.DeleteVisitByIDAndOwnerID(ctx, db.DeleteVisitByIDAndOwnerIDParams{
				ID:      ev.ID,
				OwnerID: ownerID,
			}); err != nil {
				logger.ErrorContext(ctx, "Failed to delete removed visit", "visit_id", ev.ID.String(), "error", err, "user_id", ownerID.String())
				return nil, err
			}
			deletedCount++
		}
	}

	updatedCount := 0
	createdCount := 0
	for studentID, req := range reqMap {
		if ev, ok := existingMap[studentID]; ok {
			_, err := qtx.UpdateVisit(ctx, db.UpdateVisitParams{
				ID:      ev.ID,
				OwnerID: ownerID,
				Status:  db.VisitsModuleVisitStatus(req.Status),
				Type:    db.VisitsModuleVisitType(req.Type),
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to update visit", "visit_id", ev.ID.String(), "error", err, "user_id", ownerID.String())
				return nil, err
			}
			updatedCount++
		} else {
			_, err := qtx.CreateVisit(ctx, db.CreateVisitParams{
				OwnerID:   ownerID,
				LessonID:  lessonID,
				StudentID: req.StudentID,
				Status:    db.VisitsModuleVisitStatus(req.Status),
				Type:      db.VisitsModuleVisitType(req.Type),
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to create new visit during update", "student_id", req.StudentID.String(), "error", err, "user_id", ownerID.String())
				return nil, err
			}
			createdCount++
		}
	}

	logger.DebugContext(ctx, "Visits sync details", "lesson_id", lessonID.String(), "deleted", deletedCount, "updated", updatedCount, "created", createdCount, "user_id", ownerID.String())
	return previousResponses, nil
}
