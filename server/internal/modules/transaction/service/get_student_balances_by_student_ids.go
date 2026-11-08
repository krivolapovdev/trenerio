package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) GetStudentBalancesByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID]StudentBalance, error) {
	if len(studentIDs) == 0 {
		return map[uuid.UUID]StudentBalance{}, nil
	}

	rows, err := s.db.FindStudentBalancesByStudentIDsAndOwnerID(ctx, db.FindStudentBalancesByStudentIDsAndOwnerIDParams{
		OwnerID: ownerID,
		Column2: studentIDs,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get student balances by student ids", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	balances := make(map[uuid.UUID]StudentBalance, len(rows))
	for _, r := range rows {
		balances[r.StudentID] = mapper.ToStudentBalanceFromIDsRow(r)
	}
	return balances, nil
}
