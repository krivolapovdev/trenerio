package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
)

func (s *Service) GetStudentPayments(ctx context.Context, studentID, ownerID uuid.UUID) ([]txmodel.Transaction, error) {
	logger.InfoContext(ctx, "Getting payments for student", "student_id", studentID.String(), "user_id", ownerID.String())
	if _, err := s.GetStudentDetailsByID(ctx, studentID, ownerID); err != nil {
		logger.WarnContext(ctx, "Student not found for payments", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	var payments []txmodel.Transaction
	if s.transactionProvider != nil {
		var err error
		payments, err = s.transactionProvider.GetTransactionsByStudentID(ctx, studentID, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch student payments", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return nil, err
		}
	}
	logger.InfoContext(ctx, "Retrieved student payments", "student_id", studentID.String(), "count", len(payments), "user_id", ownerID.String())
	return payments, nil
}
