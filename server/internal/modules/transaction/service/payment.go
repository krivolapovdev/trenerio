package service

import (
	"time"

	"github.com/google/uuid"
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

type AttendanceBalanceChange struct {
	StudentID     uuid.UUID
	OldBillable   bool
	OldLessonDate time.Time
	NewBillable   bool
	NewLessonDate time.Time
}

func ValidateStudentPaymentFields(paidUntil *time.Time, sessionsAdded *int64) error {
	if paidUntil == nil && sessionsAdded == nil {
		return pkgerrors.NewBadRequest("either paidUntil or sessionsAdded must be provided for student payment")
	}
	if sessionsAdded != nil && *sessionsAdded <= 0 {
		return pkgerrors.NewBadRequest("sessionsAdded must be greater than 0")
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
	if !isBillable {
		return false
	}
	if paidUntil == nil {
		return true
	}
	return IsPaidUntilBeforeDate(*paidUntil, lessonDate)
}

// ResolveAttendanceSessionBalance calculates the updated sessionsLeft balance
// for a single attendance change (recording, editing, or deleting a visit).
//
// When a visit transitions to "needs session" (e.g. marked present on a date
// not covered by paid_until), 1 session is consumed, floored at 0 (debt visit).
//
// When reverting an uncovered billable visit (edit/delete), at most 1 session is refunded,
// capped at max(0, totalSessionsAdded - remainingUncoveredBillableVisits) so debt visits
// never fabricate sessions.
func ResolveAttendanceSessionBalance(
	change AttendanceBalanceChange,
	paidUntil *time.Time,
	currentSessionsLeft int64,
	totalSessionsAdded int64,
	remainingUncoveredBillableVisits int64,
) (newSessionsLeft int64, changed bool) {
	oldNeeds := NeedsSessionForAttendance(paidUntil, change.OldLessonDate, change.OldBillable)
	newNeeds := NeedsSessionForAttendance(paidUntil, change.NewLessonDate, change.NewBillable)

	switch {
	case !oldNeeds && newNeeds:
		if currentSessionsLeft > 0 {
			return currentSessionsLeft - 1, true
		}
		return 0, false
	case oldNeeds && !newNeeds:
		maxAllowed := totalSessionsAdded - remainingUncoveredBillableVisits
		if maxAllowed < 0 {
			maxAllowed = 0
		}
		if currentSessionsLeft < maxAllowed {
			return currentSessionsLeft + 1, true
		}
		return currentSessionsLeft, false
	default:
		return currentSessionsLeft, false
	}
}
