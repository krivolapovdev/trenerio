package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/student/mapper"
)

func (s *Service) GetAllStudents(ctx context.Context, ownerID uuid.UUID) ([]studentmodel.Student, error) {
	logger.InfoContext(ctx, "Getting all students", "user_id", ownerID.String())
	rows, err := s.db.FindAllStudentsByOwnerID(ctx, ownerID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get all students", "error", err, "user_id", ownerID.String())
		return nil, err
	}
	students := make([]studentmodel.Student, 0, len(rows))
	for _, r := range rows {
		students = append(students, mapper.ToStudentModel(r))
	}
	logger.DebugContext(ctx, "Retrieved all students", "count", len(students), "user_id", ownerID.String())
	return students, nil
}
