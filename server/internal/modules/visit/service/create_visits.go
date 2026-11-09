package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func (s *Service) CreateVisits(ctx context.Context, lessonID uuid.UUID, studentVisits []StudentVisit, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Creating visits", "lesson_id", lessonID.String(), "count", len(studentVisits), "user_id", ownerID.String())
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		return s.CreateVisitsTx(ctx, qtx, lessonID, studentVisits, ownerID)
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create visits in transaction", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
		return err
	}
	logger.InfoContext(ctx, "Visits created successfully", "lesson_id", lessonID.String(), "count", len(studentVisits), "user_id", ownerID.String())
	return nil
}

func (s *Service) CreateVisitsTx(ctx context.Context, qtx *db.Queries, lessonID uuid.UUID, studentVisits []StudentVisit, ownerID uuid.UUID) error {
	for _, sv := range studentVisits {
		sv.Normalize()
		_, err := qtx.CreateVisit(ctx, db.CreateVisitParams{
			OwnerID:   ownerID,
			LessonID:  lessonID,
			StudentID: sv.StudentID,
			Status:    db.VisitsModuleVisitStatus(sv.Status),
			Type:      db.VisitsModuleVisitType(sv.Type),
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create visit record", "lesson_id", lessonID.String(), "student_id", sv.StudentID.String(), "error", err, "user_id", ownerID.String())
			return err
		}
	}
	return nil
}
