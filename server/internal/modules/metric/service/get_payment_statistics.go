package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/modules/metric/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
)

func (s *Service) GetPaymentStatistics(ctx context.Context, scopeStr string, startDate, endDate time.Time, ownerID uuid.UUID) ([]domain.PaymentMetric, error) {
	if endDate.Before(startDate) {
		logger.WarnContext(ctx, "Invalid date range for payment statistics",
			"start_date", startDate.Format(time.RFC3339),
			"end_date", endDate.Format(time.RFC3339),
			"user_id", ownerID.String(),
		)
		return nil, pkgerrors.NewBadRequest("endDate must not be before startDate")
	}

	scope := strings.ToUpper(strings.TrimSpace(scopeStr))
	if scope == "" {
		scope = string(Month)
	}

	strategy, ok := scopeStrategies[Scope(scope)]
	if !ok {
		strategy = scopeStrategies[Month]
	}

	logger.InfoContext(ctx, "Getting payment statistics",
		"scope", scope,
		"start_date", startDate.Format(time.RFC3339),
		"end_date", endDate.Format(time.RFC3339),
		"user_id", ownerID.String(),
	)

	var txList []txmodel.Transaction
	if s.transactionProvider != nil {
		var err error
		txList, err = s.transactionProvider.GetTransactionsBetween(ctx, startDate, endDate, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch transactions for payment statistics",
				"error", err,
				"start_date", startDate.Format(time.RFC3339),
				"end_date", endDate.Format(time.RFC3339),
				"user_id", ownerID.String(),
			)
			return nil, err
		}
	}

	loc := startDate.Location()
	totalsByBucket := make(map[int64]float64)
	for _, tx := range txList {
		bucket := strategy.Truncate(tx.Date, loc)
		if tx.Type == txmodel.TransactionTypeIncome {
			totalsByBucket[bucket.Unix()] += tx.Amount
		} else {
			totalsByBucket[bucket.Unix()] -= tx.Amount
		}
	}

	metrics := make([]domain.PaymentMetric, 0)
	current := strategy.Truncate(startDate, loc)
	limit := strategy.Truncate(endDate, loc)
	for !current.After(limit) {
		amount := totalsByBucket[current.Unix()]
		metrics = append(metrics, domain.PaymentMetric{
			Date:  current,
			Total: amount,
		})
		current = strategy.Next(current)
	}

	logger.InfoContext(ctx, "Payment statistics computed successfully",
		"scope", scope,
		"transactions_count", len(txList),
		"buckets_count", len(metrics),
		"user_id", ownerID.String(),
	)
	return metrics, nil
}
