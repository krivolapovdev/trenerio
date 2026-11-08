package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/student/mapper"
)

func (s *Service) GetStudentsByIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) ([]studentmodel.Student, error) {
	if len(studentIDs) == 0 {
		logger.DebugContext(ctx, "No student IDs provided to get students", "user_id", ownerID.String())
		return []studentmodel.Student{}, nil
	}

	rows, err := s.db.FindStudentsByIDsAndOwnerID(ctx, db.FindStudentsByIDsAndOwnerIDParams{
		OwnerID: ownerID,
		Column2: studentIDs,
	})

	if err != nil {
		logger.ErrorContext(ctx, "Failed to get students by IDs", "count", len(studentIDs), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	students := make([]studentmodel.Student, 0, len(rows))
	for _, r := range rows {
		students = append(students, mapper.ToStudentModel(r))
	}

	logger.DebugContext(ctx, "Retrieved students by IDs", "count", len(students), "user_id", ownerID.String())
	return students, nil
}
