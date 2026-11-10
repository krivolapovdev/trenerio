package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
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

func extractPaymentUpdateFields(updates map[string]any) map[string]any {
	merged := make(map[string]any, 3)
	if nested, ok := updates["paymentDetails"].(map[string]any); ok {
		if v, exists := nested["paidFrom"]; exists {
			merged["paidFrom"] = v
		}
		if v, exists := nested["paidUntil"]; exists {
			merged["paidUntil"] = v
		}
		if v, exists := nested["sessionsAdded"]; exists {
			merged["sessionsAdded"] = v
		}
	}
	if v, exists := updates["paidFrom"]; exists {
		merged["paidFrom"] = v
	}
	if v, exists := updates["paidUntil"]; exists {
		merged["paidUntil"] = v
	}
	if v, exists := updates["sessionsAdded"]; exists {
		merged["sessionsAdded"] = v
	}
	return merged
}

func parseAmountString(raw any) (string, error) {
	if raw == nil {
		return "", pkgerrors.NewBadRequest("amount must be greater than 0")
	}
	var val float64
	switch v := raw.(type) {
	case float64:
		val = v
	case int:
		val = float64(v)
	case int64:
		val = float64(v)
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return "", pkgerrors.NewBadRequest("invalid amount")
		}
		val = parsed
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "", pkgerrors.NewBadRequest("invalid amount")
		}
		val = parsed
	default:
		return "", pkgerrors.NewBadRequest("invalid amount")
	}
	if val <= 0 || math.IsNaN(val) || math.IsInf(val, 0) {
		return "", pkgerrors.NewBadRequest("amount must be greater than 0")
	}
	return strconv.FormatFloat(val, 'f', 2, 64), nil
}

func parsePositiveInt64(raw any) (int64, error) {
	var val int64
	switch v := raw.(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Trunc(v) != v {
			return 0, pkgerrors.NewBadRequest("sessionsAdded must be a positive integer")
		}
		val = int64(v)
	case int:
		val = int64(v)
	case int64:
		val = v
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0, pkgerrors.NewBadRequest("sessionsAdded must be a positive integer")
		}
		val = parsed
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, pkgerrors.NewBadRequest("sessionsAdded must be a positive integer")
		}
		val = parsed
	default:
		return 0, pkgerrors.NewBadRequest("sessionsAdded must be a positive integer")
	}
	if val <= 0 {
		return 0, pkgerrors.NewBadRequest("sessionsAdded must be greater than 0")
	}
	return val, nil
}

func parseTime(raw any) (time.Time, error) {
	switch val := raw.(type) {
	case string:
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			time.DateOnly,
			"2006-01-02 15:04:05",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, val); err == nil {
				return t, nil
			}
		}
		return time.Time{}, errors.New("cannot parse time")
	case time.Time:
		return val, nil
	default:
		return time.Time{}, errors.New("cannot parse time")
	}
}
