package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
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

	payments, err := s.db.FindStudentPaymentsByStudentIDsAndOwnerID(ctx, db.FindStudentPaymentsByStudentIDsAndOwnerIDParams{
		OwnerID: ownerID,
		Column2: studentIDs,
	})
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
		bal := mapper.ToStudentBalanceFromIDsRow(r)
		bal.Subscriptions = subsByStudentID[r.StudentID]
		balances[r.StudentID] = bal
	}
	return balances, nil
}
