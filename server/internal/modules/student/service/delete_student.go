package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func (s *Service) DeleteStudent(ctx context.Context, studentID, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Deleting student", "student_id", studentID.String(), "user_id", ownerID.String())
	_, err := s.db.FindStudentByIDAndOwnerID(ctx, db.FindStudentByIDAndOwnerIDParams{
		ID:      studentID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Student to delete not found", "student_id", studentID.String(), "user_id", ownerID.String())
			return pkgerrors.NewNotFound("Student", studentID, ownerID)
		}
		logger.ErrorContext(ctx, "Failed to find student for deletion", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return err
	}

	if err := s.db.DeleteStudentByIDAndOwnerID(ctx, db.DeleteStudentByIDAndOwnerIDParams{
		ID:      studentID,
		OwnerID: ownerID,
	}); err != nil {
		logger.ErrorContext(ctx, "Failed to delete student record", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return err
	}

	logger.InfoContext(ctx, "Student deleted successfully", "student_id", studentID.String(), "user_id", ownerID.String())
	return nil
}
