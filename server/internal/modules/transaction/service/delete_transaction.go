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
		existing, err := qtx.FindTransactionByIDAndOwnerID(ctx, db.FindTransactionByIDAndOwnerIDParams{
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

		var currentBal db.PaymentsModuleStudentBalance
		if existing.StudentID.Valid {
			currentBal, err = lockOrInitStudentBalanceTx(ctx, qtx, ownerID, existing.StudentID.UUID)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to lock student balance for payment deletion", "student_id", existing.StudentID.UUID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
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

		if existing.StudentID.Valid {
			studentID := existing.StudentID.UUID
			latestPaidUntil, err := findLatestPaidUntilTx(ctx, qtx, studentID, ownerID)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to recalculate latest paid_until on payment deletion", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
				return err
			}

			var oldSessions int64
			if existing.SessionsAdded.Valid {
				oldSessions = existing.SessionsAdded.Int64
			}

			updatedSessionsLeft := currentBal.SessionsLeft - oldSessions
			if updatedSessionsLeft < 0 {
				updatedSessionsLeft = 0
			}

			if _, err := qtx.UpsertStudentBalance(ctx, db.UpsertStudentBalanceParams{
				OwnerID:      ownerID,
				StudentID:    studentID,
				PaidUntil:    latestPaidUntil,
				SessionsLeft: updatedSessionsLeft,
			}); err != nil {
				logger.ErrorContext(ctx, "Failed to update student balance on payment deletion", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
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

// SyncLessonAttendanceBalancesTx applies the Priority Rule to student_balances for a batch of
// lesson attendance transitions within an existing database transaction.
