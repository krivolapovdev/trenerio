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

func (s *Service) DeleteTransaction(ctx context.Context, transactionID, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Deleting transaction", "transaction_id", transactionID.String(), "user_id", ownerID.String())
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		_, err := qtx.FindTransactionByIDAndOwnerID(ctx, db.FindTransactionByIDAndOwnerIDParams{
			ID:      transactionID,
			OwnerID: ownerID,
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Transaction to delete not found", "transaction_id", transactionID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Transaction", transactionID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find transaction for deletion", "transaction_id", transactionID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		if err := qtx.DeleteStudentPaymentByTransactionID(ctx, transactionID); err != nil {
			logger.ErrorContext(ctx, "Failed to delete student payment linked to transaction", "transaction_id", transactionID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		if err := qtx.DeleteTransactionByIDAndOwnerID(ctx, db.DeleteTransactionByIDAndOwnerIDParams{
			ID:      transactionID,
			OwnerID: ownerID,
		}); err != nil {
			logger.ErrorContext(ctx, "Failed to delete transaction record", "transaction_id", transactionID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		return nil
	})

	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete transaction from database", "transaction_id", transactionID.String(), "error", err, "user_id", ownerID.String())
		return err
	}
	logger.InfoContext(ctx, "Transaction deleted successfully", "transaction_id", transactionID.String(), "user_id", ownerID.String())
	return nil
}
