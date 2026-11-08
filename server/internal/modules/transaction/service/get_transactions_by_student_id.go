package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) GetTransactionsByStudentID(ctx context.Context, studentID, ownerID uuid.UUID) ([]domain.Transaction, error) {
	logger.InfoContext(ctx, "Getting transactions by student id", "student_id", studentID.String(), "user_id", ownerID.String())
	rows, err := s.db.FindStudentPaymentsByStudentIDAndOwnerID(ctx, db.FindStudentPaymentsByStudentIDAndOwnerIDParams{
		StudentID: studentID,
		OwnerID:   ownerID,
	})

	if err != nil {
		logger.ErrorContext(ctx, "Failed to get transactions by student id", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	transactions := make([]domain.Transaction, 0, len(rows))
	for _, r := range rows {
		transactions = append(transactions, mapper.ToTransactionFromStudentPaymentRow(r))
	}
	logger.DebugContext(ctx, "Retrieved transactions by student id", "student_id", studentID.String(), "count", len(transactions), "user_id", ownerID.String())
	return transactions, nil
}
