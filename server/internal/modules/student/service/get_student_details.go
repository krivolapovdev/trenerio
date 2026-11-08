package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/student/mapper"
)

func (s *Service) GetStudentDetailsByID(ctx context.Context, studentID, ownerID uuid.UUID) (studentmodel.Student, error) {
	logger.InfoContext(ctx, "Getting student by id", "student_id", studentID.String(), "user_id", ownerID.String())
	st, err := s.db.FindStudentByIDAndOwnerID(ctx, db.FindStudentByIDAndOwnerIDParams{
		ID:      studentID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Student not found", "student_id", studentID.String(), "user_id", ownerID.String())
			return studentmodel.Student{}, pkgerrors.NewNotFound("Student", studentID, ownerID)
		}
		logger.ErrorContext(ctx, "Failed to get student by id", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return studentmodel.Student{}, err
	}
	logger.DebugContext(ctx, "Retrieved student by id", "student_id", studentID.String(), "user_id", ownerID.String())
	return mapper.ToStudentModel(st), nil
}

func (s *Service) ValidateStudentExistsByID(ctx context.Context, studentID, ownerID uuid.UUID) error {
	_, err := s.GetStudentDetailsByID(ctx, studentID, ownerID)
	return err
}
