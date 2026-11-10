package service

import (
	"time"

	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
)

type StudentBalance = domain.StudentBalance

func IsPaidUntilBeforeDate(paidUntil, date time.Time) bool {
	return domain.IsPaidUntilBeforeDate(paidUntil, date)
}

func IsSubscriptionActive(maxPaidUntil *time.Time, referenceDate time.Time) bool {
	return domain.IsSubscriptionActive(maxPaidUntil, referenceDate)
}

func IsDateRangeActive(paidFrom, paidUntil *time.Time, date time.Time) bool {
	return domain.IsDateRangeActive(paidFrom, paidUntil, date)
}

func ValidateStudentPaymentFields(paidUntil *time.Time, sessionsAdded *int64) error {
	return ValidateStudentPaymentFieldsWithFrom(nil, paidUntil, sessionsAdded)
}

func ValidateStudentPaymentFieldsWithFrom(paidFrom, paidUntil *time.Time, sessionsAdded *int64) error {
	if paidUntil == nil && sessionsAdded == nil {
		return pkgerrors.NewBadRequest("either paidUntil or sessionsAdded must be provided for student payment")
	}
	if sessionsAdded != nil && *sessionsAdded <= 0 {
		return pkgerrors.NewBadRequest("sessionsAdded must be greater than 0")
	}
	if paidFrom != nil && paidUntil != nil && paidFrom.After(*paidUntil) {
		return pkgerrors.NewBadRequest("paidFrom cannot be after paidUntil")
	}
	return nil
}

// MaxPaidUntil extracts the latest PaidUntil date from a list of transactions.
// Transactions without PaymentDetails or PaidUntil are ignored.
// If no valid PaidUntil date is found, nil is returned.
func MaxPaidUntil(payments []domain.Transaction) *time.Time {
	var maxPaidUntil *time.Time
	for _, p := range payments {
		if p.PaymentDetails != nil && p.PaymentDetails.PaidUntil != nil {
			pu := *p.PaymentDetails.PaidUntil
			if maxPaidUntil == nil || pu.After(*maxPaidUntil) {
				t := pu
				maxPaidUntil = &t
			}
		}
	}
	return maxPaidUntil
}

func NeedsSessionForAttendance(paidUntil *time.Time, lessonDate time.Time, isBillable bool) bool {
	return NeedsSessionForAttendanceRange(nil, paidUntil, lessonDate, isBillable)
}

func NeedsSessionForAttendanceRange(paidFrom, paidUntil *time.Time, lessonDate time.Time, isBillable bool) bool {
	if !isBillable {
		return false
	}
	return !domain.IsDateRangeActive(paidFrom, paidUntil, lessonDate)
}
