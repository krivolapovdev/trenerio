package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/mapper"
)

func (s *Service) GetVisitsByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID][]VisitResponse, error) {
	logger.InfoContext(ctx, "Getting visits by student ids", "count", len(studentIDs), "user_id", ownerID.String())
	if len(studentIDs) == 0 {
		logger.DebugContext(ctx, "No student IDs provided to get visits", "user_id", ownerID.String())
		return map[uuid.UUID][]VisitResponse{}, nil
	}

	rows, err := s.db.FindVisitsByStudentIDsAndOwnerID(ctx, db.FindVisitsByStudentIDsAndOwnerIDParams{
		OwnerID: ownerID,
		Column2: studentIDs,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get visits by student ids", "count", len(studentIDs), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	visitsByStudent := make(map[uuid.UUID][]VisitResponse, len(studentIDs))
	for _, sid := range studentIDs {
		visitsByStudent[sid] = []VisitResponse{}
	}
	for _, r := range rows {
		visitsByStudent[r.StudentID] = append(visitsByStudent[r.StudentID], mapper.ToVisitResponse(r))
	}
	logger.DebugContext(ctx, "Retrieved visits by student ids", "student_count", len(studentIDs), "total_visits", len(rows), "user_id", ownerID.String())
	return visitsByStudent, nil
}
