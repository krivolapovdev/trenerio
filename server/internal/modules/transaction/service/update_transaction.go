package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) UpdateTransaction(ctx context.Context, transactionID, ownerID uuid.UUID, updates map[string]any) (domain.Transaction, error) {
	logger.InfoContext(ctx, "Updating transaction", "transaction_id", transactionID.String(), "user_id", ownerID.String())

	var patchedAmountStr *string
	if rawAmt, ok := updates["amount"]; ok {
		amtStr, err := parseAmountString(rawAmt)
		if err != nil {
			return domain.Transaction{}, err
		}
		patchedAmountStr = &amtStr
	}

	var patchedDate *time.Time
	if rawDate, ok := updates["date"]; ok {
		if rawDate == nil {
			return domain.Transaction{}, pkgerrors.NewBadRequest("date must not be null")
		}
		d, err := parseTime(rawDate)
		if err != nil {
			return domain.Transaction{}, pkgerrors.NewBadRequest("invalid date format")
		}
		if err := validator.ValidateDateNotFuture(d, "date"); err != nil {
			logger.WarnContext(ctx, "Updated transaction date cannot be in the future", "date", d.Format(time.RFC3339), "user_id", ownerID.String())
			return domain.Transaction{}, err
		}
		patchedDate = &d
	}

	paymentUpdates := extractPaymentUpdateFields(updates)

	hasPaidFromUpdate := false
	var patchedPaidFrom sql.NullTime
	if rawPF, ok := paymentUpdates["paidFrom"]; ok {
		hasPaidFromUpdate = true
		if rawPF != nil {
			d, err := parseTime(rawPF)
			if err != nil {
				return domain.Transaction{}, pkgerrors.NewBadRequest("invalid paidFrom format")
			}
			patchedPaidFrom = sql.NullTime{Time: d, Valid: true}
		}
	}

	hasPaidUntilUpdate := false
	var patchedPaidUntil sql.NullTime
	if rawPU, ok := paymentUpdates["paidUntil"]; ok {
		hasPaidUntilUpdate = true
		if rawPU != nil {
			d, err := parseTime(rawPU)
			if err != nil {
				return domain.Transaction{}, pkgerrors.NewBadRequest("invalid paidUntil format")
			}
			patchedPaidUntil = sql.NullTime{Time: d, Valid: true}
		}
	}

	hasSessionsAddedUpdate := false
	var patchedSessionsAdded sql.NullInt64
	if rawSA, ok := paymentUpdates["sessionsAdded"]; ok {
		hasSessionsAddedUpdate = true
		if rawSA != nil {
			sa, err := parsePositiveInt64(rawSA)
			if err != nil {
				return domain.Transaction{}, err
			}
			patchedSessionsAdded = sql.NullInt64{Int64: sa, Valid: true}
		}
	}

	var updatedTx db.FinanceModuleTransaction
	var details *domain.StudentPaymentDetails

	err := s.withTx(ctx, func(qtx *db.Queries) error {
		existing, err := qtx.FindTransactionByIDAndOwnerID(ctx, db.FindTransactionByIDAndOwnerIDParams{
			ID:      transactionID,
			OwnerID: ownerID,
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Transaction to update not found", "transaction_id", transactionID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Transaction", transactionID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find transaction for update", "transaction_id", transactionID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		amountStr := existing.Amount
		if patchedAmountStr != nil {
			amountStr = *patchedAmountStr
		}

		txDate := existing.Date
		if patchedDate != nil {
			txDate = *patchedDate
		}

		newPaidFrom := existing.PaidFrom
		if hasPaidFromUpdate {
			newPaidFrom = patchedPaidFrom
		}

		newPaidUntil := existing.PaidUntil
		if hasPaidUntilUpdate {
			newPaidUntil = patchedPaidUntil
		}

		newSessionsAdded := existing.SessionsAdded
		if hasSessionsAddedUpdate {
			newSessionsAdded = patchedSessionsAdded
		}

		patchesPayment := hasPaidFromUpdate || hasPaidUntilUpdate || hasSessionsAddedUpdate
		if existing.StudentID.Valid && patchesPayment && !newPaidUntil.Valid && !newSessionsAdded.Valid {
			return pkgerrors.NewBadRequest("either paidUntil or sessionsAdded must be provided for student payment")
		}

		if newPaidFrom.Valid && newPaidUntil.Valid && newPaidFrom.Time.After(newPaidUntil.Time) {
			return pkgerrors.NewBadRequest("paidFrom cannot be after paidUntil")
		}

		updatedTx, err = qtx.UpdateTransaction(ctx, db.UpdateTransactionParams{
			ID:      transactionID,
			OwnerID: ownerID,
			Amount:  amountStr,
			Date:    txDate,
		})

		if err != nil {
			logger.ErrorContext(ctx, "Failed to update transaction record", "transaction_id", transactionID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		if patchesPayment && existing.StudentID.Valid {
			sp, err := qtx.UpdateStudentPayment(ctx, db.UpdateStudentPaymentParams{
				TransactionID: transactionID,
				PaidFrom:      newPaidFrom,
				PaidUntil:     newPaidUntil,
				SessionsAdded: newSessionsAdded,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to update student payment record", "transaction_id", transactionID.String(), "error", err, "user_id", ownerID.String())
				return err
			}

			details = mapper.ToStudentPaymentDetails(sp.StudentID, sp.PaidFrom, sp.PaidUntil, sp.SessionsAdded, nil)
		} else if existing.StudentID.Valid {
			details = mapper.ToStudentPaymentDetails(existing.StudentID.UUID, existing.PaidFrom, existing.PaidUntil, existing.SessionsAdded, nil)
		}

		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update transaction in database", "transaction_id", transactionID.String(), "error", err, "user_id", ownerID.String())
		return domain.Transaction{}, err
	}

	amt, _ := strconv.ParseFloat(updatedTx.Amount, 64)
	logger.InfoContext(ctx, "Transaction updated successfully", "transaction_id", updatedTx.ID.String(), "amount", amt, "user_id", ownerID.String())
	return mapper.ToTransaction(updatedTx, details), nil
}
