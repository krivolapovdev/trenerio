package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) GetStudentBalancesByOwnerID(ctx context.Context, ownerID uuid.UUID) (map[uuid.UUID]StudentBalance, error) {
	logger.DebugContext(ctx, "Getting student balances by owner id", "user_id", ownerID.String())
	rows, err := s.db.FindStudentBalancesByOwnerID(ctx, ownerID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get student balances by owner id", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	payments, err := s.db.FindStudentPaymentsByOwnerID(ctx, ownerID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get student payments for subscription ranges", "error", err, "user_id", ownerID.String())
	}

	subsByStudentID := make(map[uuid.UUID][]domain.SubscriptionRange)
	for _, p := range payments {
		if p.PaidUntil.Valid {
			subsByStudentID[p.StudentID] = append(subsByStudentID[p.StudentID], domain.SubscriptionRange{
				PaidFrom:  mapper.FromNullTime(p.PaidFrom),
				PaidUntil: mapper.FromNullTime(p.PaidUntil),
			})
		}
	}

	balances := make(map[uuid.UUID]StudentBalance, len(rows))
	for _, r := range rows {
		bal := mapper.ToStudentBalanceFromOwnerRow(r)
		bal.Subscriptions = subsByStudentID[r.StudentID]
		balances[r.StudentID] = bal
	}
	return balances, nil
}
