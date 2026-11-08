package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxPaidUntil(t *testing.T) {
	t.Run("EmptyList_ReturnsNil", func(t *testing.T) {
		assert.Nil(t, MaxPaidUntil(nil))
		assert.Nil(t, MaxPaidUntil([]domain.Transaction{}))
	})

	t.Run("NoPaymentDetailsOrSessionOnly_ReturnsNil", func(t *testing.T) {
		sessions := int64(8)
		txs := []domain.Transaction{
			{PaymentDetails: nil},
			{PaymentDetails: &domain.StudentPaymentDetails{PaidUntil: nil, SessionsAdded: &sessions}},
		}
		assert.Nil(t, MaxPaidUntil(txs))
	})

	t.Run("MultiplePayments_ReturnsLatest", func(t *testing.T) {
		t1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
		sessions := int64(4)

		txs := []domain.Transaction{
			{PaymentDetails: &domain.StudentPaymentDetails{PaidUntil: &t1}},
			{PaymentDetails: nil},
			{PaymentDetails: &domain.StudentPaymentDetails{PaidUntil: nil, SessionsAdded: &sessions}},
			{PaymentDetails: &domain.StudentPaymentDetails{PaidUntil: &t2, SessionsAdded: &sessions}},
			{PaymentDetails: &domain.StudentPaymentDetails{PaidUntil: &t3}},
		}

		latest := MaxPaidUntil(txs)
		require.NotNil(t, latest)
		assert.Equal(t, t2, *latest)
	})
}

func TestValidateStudentPaymentFields(t *testing.T) {
	paidUntil := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	validSessions := int64(8)
	zeroSessions := int64(0)
	negSessions := int64(-5)

	t.Run("ValidPaidUntilOnly", func(t *testing.T) {
		assert.NoError(t, ValidateStudentPaymentFields(&paidUntil, nil))
	})

	t.Run("ValidSessionsOnly", func(t *testing.T) {
		assert.NoError(t, ValidateStudentPaymentFields(nil, &validSessions))
	})

	t.Run("ValidBoth", func(t *testing.T) {
		assert.NoError(t, ValidateStudentPaymentFields(&paidUntil, &validSessions))
	})

	t.Run("Neither_ReturnsError", func(t *testing.T) {
		assert.Error(t, ValidateStudentPaymentFields(nil, nil))
	})

	t.Run("ZeroSessions_ReturnsError", func(t *testing.T) {
		assert.Error(t, ValidateStudentPaymentFields(nil, &zeroSessions))
	})

	t.Run("NegativeSessions_ReturnsError", func(t *testing.T) {
		assert.Error(t, ValidateStudentPaymentFields(nil, &negSessions))
	})
}

func TestNeedsSessionForAttendance(t *testing.T) {
	paidUntil := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)

	t.Run("NonBillable_DoesNotNeedSession", func(t *testing.T) {
		assert.False(t, NeedsSessionForAttendance(&paidUntil, time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC), false))
		assert.False(t, NeedsSessionForAttendance(nil, time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC), false))
	})

	t.Run("NoSubscription_Billable_NeedsSession", func(t *testing.T) {
		assert.True(t, NeedsSessionForAttendance(nil, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), true))
	})

	t.Run("LessonWithinSubscription_DoesNotNeedSession", func(t *testing.T) {
		assert.False(t, NeedsSessionForAttendance(&paidUntil, time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC), true))
		assert.False(t, NeedsSessionForAttendance(&paidUntil, time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC), true))
	})

	t.Run("LessonAfterSubscription_NeedsSession", func(t *testing.T) {
		assert.True(t, NeedsSessionForAttendance(&paidUntil, time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC), true))
	})
}

func TestResolveAttendanceSessionBalance(t *testing.T) {
	paidUntil := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	coveredDate := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
	uncoveredDate := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	sID := uuid.New()

	t.Run("UncoveredMarked_DeductsSession", func(t *testing.T) {
		change := AttendanceBalanceChange{
			StudentID:     sID,
			OldBillable:   false,
			OldLessonDate: coveredDate,
			NewBillable:   true,
			NewLessonDate: uncoveredDate,
		}
		newLeft, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 5, 10, 0)
		assert.True(t, changed)
		assert.Equal(t, int64(4), newLeft)
	})

	t.Run("UncoveredMarked_ZeroBalanceFloorsAtZero", func(t *testing.T) {
		change := AttendanceBalanceChange{
			StudentID:     sID,
			OldBillable:   false,
			OldLessonDate: coveredDate,
			NewBillable:   true,
			NewLessonDate: uncoveredDate,
		}
		newLeft, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 0, 10, 0)
		assert.False(t, changed)
		assert.Equal(t, int64(0), newLeft)
	})

	t.Run("UncoveredUnmarked_RefundsSessionWhenAllowed", func(t *testing.T) {
		change := AttendanceBalanceChange{
			StudentID:     sID,
			OldBillable:   true,
			OldLessonDate: uncoveredDate,
			NewBillable:   false,
			NewLessonDate: uncoveredDate,
		}
		newLeft, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 3, 10, 2)
		assert.True(t, changed)
		assert.Equal(t, int64(4), newLeft)
	})

	t.Run("UncoveredUnmarked_DoesNotRefundDebtVisits", func(t *testing.T) {
		change := AttendanceBalanceChange{
			StudentID:     sID,
			OldBillable:   true,
			OldLessonDate: uncoveredDate,
			NewBillable:   false,
			NewLessonDate: uncoveredDate,
		}
		newLeft, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 0, 5, 5)
		assert.False(t, changed)
		assert.Equal(t, int64(0), newLeft)
	})

	t.Run("CoveredAttendance_NoBalanceChange", func(t *testing.T) {
		change := AttendanceBalanceChange{
			StudentID:     sID,
			OldBillable:   false,
			OldLessonDate: coveredDate,
			NewBillable:   true,
			NewLessonDate: coveredDate,
		}
		newLeft, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 5, 10, 0)
		assert.False(t, changed)
		assert.Equal(t, int64(5), newLeft)
	})
}
