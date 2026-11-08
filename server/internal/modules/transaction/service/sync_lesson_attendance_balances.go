package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) SyncLessonAttendanceBalancesTx(ctx context.Context, qtx *db.Queries, ownerID uuid.UUID, changes []AttendanceBalanceChange) error {
	relevant := make([]AttendanceBalanceChange, 0, len(changes))
	for _, ch := range changes {
		if ch.OldBillable || ch.NewBillable {
			relevant = append(relevant, ch)
		}
	}
	if len(relevant) == 0 {
		return nil
	}

	// Acquire row locks in deterministic student ID order to prevent deadlocks.
	slices.SortFunc(relevant, func(a, b AttendanceBalanceChange) int {
		return strings.Compare(a.StudentID.String(), b.StudentID.String())
	})

	for _, ch := range relevant {
		bal, err := lockOrInitStudentBalanceTx(ctx, qtx, ownerID, ch.StudentID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to lock student balance during attendance sync", "student_id", ch.StudentID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		paidUntil := mapper.FromNullTime(bal.PaidUntil)
		oldNeeds := NeedsSessionForAttendance(paidUntil, ch.OldLessonDate, ch.OldBillable)
		newNeeds := NeedsSessionForAttendance(paidUntil, ch.NewLessonDate, ch.NewBillable)
		if oldNeeds == newNeeds {
			continue
		}

		var totalSessionsAdded int64
		var remainingUncovered int64
		if oldNeeds && !newNeeds {
			totalSessionsAdded, err = qtx.SumSessionsAddedByStudentIDAndOwnerID(ctx, db.SumSessionsAddedByStudentIDAndOwnerIDParams{
				StudentID: ch.StudentID,
				OwnerID:   ownerID,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to sum sessions_added during attendance refund check", "student_id", ch.StudentID.String(), "error", err, "user_id", ownerID.String())
				return err
			}

			billableDates, err := qtx.FindBillableVisitLessonDatesByStudentIDAndOwnerID(ctx, db.FindBillableVisitLessonDatesByStudentIDAndOwnerIDParams{
				StudentID: ch.StudentID,
				OwnerID:   ownerID,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to query remaining billable visits during attendance refund check", "student_id", ch.StudentID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
			for _, d := range billableDates {
				if !IsSubscriptionActive(paidUntil, d) {
					remainingUncovered++
				}
			}
		}

		newSessionsLeft, changed := ResolveAttendanceSessionBalance(
			ch,
			paidUntil,
			bal.SessionsLeft,
			totalSessionsAdded,
			remainingUncovered,
		)
		if !changed {
			continue
		}

		if _, err := qtx.UpsertStudentBalance(ctx, db.UpsertStudentBalanceParams{
			OwnerID:      ownerID,
			StudentID:    ch.StudentID,
			PaidUntil:    bal.PaidUntil,
			SessionsLeft: newSessionsLeft,
		}); err != nil {
			logger.ErrorContext(ctx, "Failed to update student balance during attendance sync", "student_id", ch.StudentID.String(), "error", err, "user_id", ownerID.String())
			return err
		}
	}

	return nil
}

func findLatestPaidUntilTx(ctx context.Context, qtx *db.Queries, studentID, ownerID uuid.UUID) (sql.NullTime, error) {
	pu, err := qtx.FindLatestPaidUntilByStudentID(ctx, db.FindLatestPaidUntilByStudentIDParams{
		StudentID: studentID,
		OwnerID:   ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.NullTime{}, nil
		}
		return sql.NullTime{}, err
	}
	return pu, nil
}

func lockOrInitStudentBalanceTx(ctx context.Context, qtx *db.Queries, ownerID, studentID uuid.UUID) (db.PaymentsModuleStudentBalance, error) {
	bal, err := qtx.FindStudentBalanceForUpdate(ctx, db.FindStudentBalanceForUpdateParams{
		OwnerID:   ownerID,
		StudentID: studentID,
	})
	if err == nil {
		return bal, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return db.PaymentsModuleStudentBalance{}, err
	}

	latestPaidUntil, err := findLatestPaidUntilTx(ctx, qtx, studentID, ownerID)
	if err != nil {
		return db.PaymentsModuleStudentBalance{}, err
	}

	return qtx.UpsertStudentBalance(ctx, db.UpsertStudentBalanceParams{
		OwnerID:      ownerID,
		StudentID:    studentID,
		PaidUntil:    latestPaidUntil,
		SessionsLeft: 0,
	})
}

func extractPaymentUpdateFields(updates map[string]any) map[string]any {
	merged := make(map[string]any, 2)
	if nested, ok := updates["paymentDetails"].(map[string]any); ok {
		if v, exists := nested["paidUntil"]; exists {
			merged["paidUntil"] = v
		}
		if v, exists := nested["sessionsAdded"]; exists {
			merged["sessionsAdded"] = v
		}
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
