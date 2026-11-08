package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) GetStudentBalancesByOwnerID(ctx context.Context, ownerID uuid.UUID) (map[uuid.UUID]StudentBalance, error) {
	logger.DebugContext(ctx, "Getting student balances by owner id", "user_id", ownerID.String())
	rows, err := s.db.FindStudentBalancesByOwnerID(ctx, ownerID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get student balances by owner id", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	balances := make(map[uuid.UUID]StudentBalance, len(rows))
	for _, r := range rows {
		balances[r.StudentID] = mapper.ToStudentBalanceFromOwnerRow(r)
	}
	return balances, nil
}
