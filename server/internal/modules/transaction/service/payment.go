package service

import (
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/model"
)

type StudentBalance = model.StudentBalance

func IsPaidUntilBeforeDate(paidUntil, date time.Time) bool {
	return model.IsPaidUntilBeforeDate(paidUntil, date)
}

func IsSubscriptionActive(maxPaidUntil *time.Time, referenceDate time.Time) bool {
	return model.IsSubscriptionActive(maxPaidUntil, referenceDate)
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
func MaxPaidUntil(payments []response.TransactionResponse) *time.Time {
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

// NeedsSessionForAttendance returns true under the Priority Rule when a visit is billable
// and is not covered by an active subscription on lessonDate.
func NeedsSessionForAttendance(paidUntil *time.Time, lessonDate time.Time, billable bool) bool {
	if !billable {
		return false
	}
	return !IsSubscriptionActive(paidUntil, lessonDate)
}

// ResolveAttendanceSessionBalance applies the Priority Rule for an attendance transition:
//  1. Date Check: If lessonDate <= paidUntil, sessions_left is NOT modified.
//  2. Session Check: If subscription is inactive/null and sessions_left > 0, decrement by 1.
//  3. Debt: If subscription is inactive/null and sessions_left == 0, sessions_left stays 0.
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
