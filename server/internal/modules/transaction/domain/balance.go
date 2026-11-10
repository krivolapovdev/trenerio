package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/clock"
)

// @name SubscriptionRange
type SubscriptionRange struct {
	PaidFrom  *time.Time
	PaidUntil *time.Time
}

// @name StudentBalance
type StudentBalance struct {
	StudentID          uuid.UUID
	PaidFrom           *time.Time
	PaidUntil          *time.Time
	Subscriptions      []SubscriptionRange
	SessionsLeft       int64
	TotalSessionsAdded int64
}

func (b *StudentBalance) HasPaymentActivity() bool {
	if b == nil {
		return false
	}
	return b.PaidUntil != nil || b.PaidFrom != nil || b.SessionsLeft > 0 || b.TotalSessionsAdded > 0
}

func (b *StudentBalance) IsCoveredByDateRange(date time.Time) bool {
	if b == nil {
		return false
	}
	if len(b.Subscriptions) > 0 {
		for _, s := range b.Subscriptions {
			if IsDateRangeActive(s.PaidFrom, s.PaidUntil, date) {
				return true
			}
		}
		return false
	}
	return IsDateRangeActive(b.PaidFrom, b.PaidUntil, date)
}

func (b *StudentBalance) IsPaidAt(referenceDate time.Time) bool {
	if b == nil {
		return false
	}
	return b.IsCoveredByDateRange(referenceDate) || b.SessionsLeft > 0
}

func IsDateRangeActive(paidFrom, paidUntil *time.Time, date time.Time) bool {
	if paidUntil == nil {
		return false
	}
	if clock.StartOfDay(date).After(clock.EndOfDay(*paidUntil)) {
		if clock.StartOfDay(date.UTC()).After(clock.EndOfDay(paidUntil.UTC())) {
			return false
		}
	}
	if paidFrom != nil {
		if clock.EndOfDay(date).Before(clock.StartOfDay(*paidFrom)) {
			if clock.EndOfDay(date.UTC()).Before(clock.StartOfDay(paidFrom.UTC())) {
				return false
			}
		}
	}
	return true
}

func IsPaidUntilBeforeDate(paidUntil, date time.Time) bool {
	paidUntilEnd := clock.EndOfDay(paidUntil)
	dateStart := clock.StartOfDay(date)

	if paidUntilEnd.Before(dateStart) {
		paidUntilEndUTC := clock.EndOfDay(paidUntil.UTC())
		dateStartUTC := clock.StartOfDay(date.UTC())
		return paidUntilEndUTC.Before(dateStartUTC)
	}

	return false
}

func IsSubscriptionActive(maxPaidUntil *time.Time, referenceDate time.Time) bool {
	if maxPaidUntil == nil {
		return false
	}
	return !IsPaidUntilBeforeDate(*maxPaidUntil, referenceDate)
}
