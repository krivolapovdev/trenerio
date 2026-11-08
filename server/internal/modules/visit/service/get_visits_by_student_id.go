package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/mapper"
)

func (s *Service) GetVisitsByStudentID(ctx context.Context, studentID, ownerID uuid.UUID) ([]VisitResponse, error) {
	logger.InfoContext(ctx, "Getting visits by student id", "student_id", studentID.String(), "user_id", ownerID.String())
	rows, err := s.db.FindVisitsByStudentIDAndOwnerID(ctx, db.FindVisitsByStudentIDAndOwnerIDParams{
		StudentID: studentID,
		OwnerID:   ownerID,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get visits by student id", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	visits := make([]VisitResponse, 0, len(rows))
	for _, r := range rows {
		visits = append(visits, mapper.ToVisitResponse(r))
	}
	logger.DebugContext(ctx, "Retrieved visits by student id", "student_id", studentID.String(), "count", len(visits), "user_id", ownerID.String())
	return visits, nil
}
