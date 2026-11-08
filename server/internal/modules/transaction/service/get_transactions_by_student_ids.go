package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) GetTransactionsByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID][]domain.Transaction, error) {
	logger.InfoContext(ctx, "Getting transactions by student ids", "count", len(studentIDs), "user_id", ownerID.String())
	if len(studentIDs) == 0 {
		logger.DebugContext(ctx, "No student IDs provided to get transactions", "user_id", ownerID.String())
		return map[uuid.UUID][]domain.Transaction{}, nil
	}

	rows, err := s.db.FindStudentPaymentsByStudentIDsAndOwnerID(ctx, db.FindStudentPaymentsByStudentIDsAndOwnerIDParams{
		OwnerID: ownerID,
		Column2: studentIDs,
	})

	if err != nil {
		logger.ErrorContext(ctx, "Failed to get transactions by student ids", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	transactionsByStudent := make(map[uuid.UUID][]domain.Transaction, len(studentIDs))
	for _, sid := range studentIDs {
		transactionsByStudent[sid] = []domain.Transaction{}
	}

	for _, r := range rows {
		transactionsByStudent[r.StudentID] = append(transactionsByStudent[r.StudentID], mapper.ToTransactionFromStudentPaymentsRow(r))
	}

	logger.DebugContext(ctx, "Retrieved transactions by student ids", "count", len(rows), "user_id", ownerID.String())
	return transactionsByStudent, nil
}
