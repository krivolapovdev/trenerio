package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/clock"
)

type StudentBalance struct {
	StudentID          uuid.UUID
	PaidUntil          *time.Time
	SessionsLeft       int64
	TotalSessionsAdded int64
}

func (b *StudentBalance) HasPaymentActivity() bool {
	if b == nil {
		return false
	}
	return b.PaidUntil != nil || b.SessionsLeft > 0 || b.TotalSessionsAdded > 0
}

func (b *StudentBalance) IsPaidAt(referenceDate time.Time) bool {
	if b == nil {
		return false
	}
	return IsSubscriptionActive(b.PaidUntil, referenceDate) || b.SessionsLeft > 0
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
