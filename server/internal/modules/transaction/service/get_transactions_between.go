package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) GetTransactionsBetween(ctx context.Context, startDate, endDate time.Time, ownerID uuid.UUID) ([]domain.Transaction, error) {
	logger.InfoContext(ctx, "Getting transactions between dates", "start_date", startDate.Format(time.RFC3339), "end_date", endDate.Format(time.RFC3339), "user_id", ownerID.String())
	rows, err := s.db.FindTransactionsByOwnerIDAndDateBetween(ctx, db.FindTransactionsByOwnerIDAndDateBetweenParams{
		OwnerID: ownerID,
		Date:    startDate,
		Date_2:  endDate,
	})

	if err != nil {
		logger.ErrorContext(ctx, "Failed to get transactions between dates", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	transactions := make([]domain.Transaction, 0, len(rows))
	for _, r := range rows {
		transactions = append(transactions, mapper.ToTransactionFromDateBetweenRow(r))
	}

	logger.DebugContext(ctx, "Retrieved transactions between dates", "count", len(transactions), "user_id", ownerID.String())
	return transactions, nil
}
