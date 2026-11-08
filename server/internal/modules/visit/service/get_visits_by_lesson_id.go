package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/mapper"
)

func (s *Service) GetVisitsByLessonID(ctx context.Context, lessonID, ownerID uuid.UUID) ([]VisitResponse, error) {
	return s.GetVisitsByLessonIDTx(ctx, s.db, lessonID, ownerID)
}

func (s *Service) GetVisitsByLessonIDTx(ctx context.Context, qtx *db.Queries, lessonID, ownerID uuid.UUID) ([]VisitResponse, error) {
	logger.InfoContext(ctx, "Getting visits by lesson id", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	rows, err := qtx.FindVisitsByLessonIDAndOwnerID(ctx, db.FindVisitsByLessonIDAndOwnerIDParams{
		LessonID: lessonID,
		OwnerID:  ownerID,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get visits by lesson id", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	visits := make([]VisitResponse, 0, len(rows))
	for _, r := range rows {
		visits = append(visits, mapper.ToVisitResponse(r))
	}
	logger.DebugContext(ctx, "Retrieved visits by lesson id", "lesson_id", lessonID.String(), "count", len(visits), "user_id", ownerID.String())
	return visits, nil
}
