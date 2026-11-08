package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/mapper"
)

func (s *Service) GetVisitsByLessonIDs(ctx context.Context, lessonIDs []uuid.UUID, ownerID uuid.UUID) ([]VisitResponse, error) {
	logger.InfoContext(ctx, "Getting visits by lesson ids", "count", len(lessonIDs), "user_id", ownerID.String())
	if len(lessonIDs) == 0 {
		logger.DebugContext(ctx, "No lesson IDs provided to get visits", "user_id", ownerID.String())
		return []VisitResponse{}, nil
	}

	rows, err := s.db.FindVisitsByLessonIDsAndOwnerID(ctx, db.FindVisitsByLessonIDsAndOwnerIDParams{
		OwnerID: ownerID,
		Column2: lessonIDs,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get visits by lesson ids", "count", len(lessonIDs), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	visits := make([]VisitResponse, 0, len(rows))
	for _, r := range rows {
		visits = append(visits, mapper.ToVisitResponse(r))
	}
	logger.DebugContext(ctx, "Retrieved visits by lesson ids", "count", len(visits), "user_id", ownerID.String())
	return visits, nil
}
