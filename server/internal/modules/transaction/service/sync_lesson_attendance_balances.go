package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/db"
)

// SyncLessonAttendanceBalancesTx is a no-op now that student balances are computed dynamically.
// Kept for backward compatibility with the lesson service balanceProvider interface.
func (s *Service) SyncLessonAttendanceBalancesTx(ctx context.Context, qtx *db.Queries, ownerID uuid.UUID, changes []AttendanceBalanceChange) error {
	return nil
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
