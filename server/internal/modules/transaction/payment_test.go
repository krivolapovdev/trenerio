package transaction

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxPaidUntil(t *testing.T) {
	t.Run("EmptyList_ReturnsNil", func(t *testing.T) {
		assert.Nil(t, MaxPaidUntil(nil))
		assert.Nil(t, MaxPaidUntil([]TransactionResponse{}))
	})

	t.Run("NoPaymentDetailsOrSessionOnly_ReturnsNil", func(t *testing.T) {
		sessions := int64(8)
		txs := []TransactionResponse{
			{PaymentDetails: nil},
			{PaymentDetails: &StudentPaymentDetailsResponse{PaidUntil: nil, SessionsAdded: &sessions}},
		}
		assert.Nil(t, MaxPaidUntil(txs))
	})

	t.Run("MultiplePayments_ReturnsLatest", func(t *testing.T) {
		t1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
		sessions := int64(4)

		txs := []TransactionResponse{
			{PaymentDetails: &StudentPaymentDetailsResponse{PaidUntil: &t1}},
			{PaymentDetails: nil},
			{PaymentDetails: &StudentPaymentDetailsResponse{PaidUntil: nil, SessionsAdded: &sessions}},
			{PaymentDetails: &StudentPaymentDetailsResponse{PaidUntil: &t2, SessionsAdded: &sessions}},
			{PaymentDetails: &StudentPaymentDetailsResponse{PaidUntil: &t3}},
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
	negSessions := int64(-3)

	assert.NoError(t, ValidateStudentPaymentFields(&paidUntil, nil))
	assert.NoError(t, ValidateStudentPaymentFields(nil, &validSessions))
	assert.NoError(t, ValidateStudentPaymentFields(&paidUntil, &validSessions))

	assert.Error(t, ValidateStudentPaymentFields(nil, nil))
	assert.Error(t, ValidateStudentPaymentFields(nil, &zeroSessions))
	assert.Error(t, ValidateStudentPaymentFields(&paidUntil, &negSessions))
}

func TestStudentBalanceHelpers(t *testing.T) {
	refDate := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	activeDate := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	expiredDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	var nilBal *StudentBalance
	assert.False(t, nilBal.HasPaymentActivity())
	assert.False(t, nilBal.IsPaidAt(refDate))

	emptyBal := &StudentBalance{StudentID: uuid.New(), PaidUntil: nil, SessionsLeft: 0}
	assert.False(t, emptyBal.HasPaymentActivity())
	assert.False(t, emptyBal.IsPaidAt(refDate))

	sessionOnlyBal := &StudentBalance{StudentID: uuid.New(), PaidUntil: nil, SessionsLeft: 3}
	assert.True(t, sessionOnlyBal.HasPaymentActivity())
	assert.True(t, sessionOnlyBal.IsPaidAt(refDate))

	expiredWithReserve := &StudentBalance{StudentID: uuid.New(), PaidUntil: &expiredDate, SessionsLeft: 2}
	assert.True(t, expiredWithReserve.HasPaymentActivity())
	assert.True(t, expiredWithReserve.IsPaidAt(refDate))

	expiredWithoutReserve := &StudentBalance{StudentID: uuid.New(), PaidUntil: &expiredDate, SessionsLeft: 0}
	assert.True(t, expiredWithoutReserve.HasPaymentActivity())
	assert.False(t, expiredWithoutReserve.IsPaidAt(refDate))

	activeSub := &StudentBalance{StudentID: uuid.New(), PaidUntil: &activeDate, SessionsLeft: 0}
	assert.True(t, activeSub.HasPaymentActivity())
	assert.True(t, activeSub.IsPaidAt(refDate))
}

func TestResolveAttendanceSessionBalance_PriorityRule(t *testing.T) {
	studentID := uuid.New()
	paidUntil := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

	t.Run("Priority1_DateActive_DoesNotDeductReserveSessions", func(t *testing.T) {
		lessonOnPaidUntil := time.Date(2026, 11, 15, 18, 0, 0, 0, time.UTC)
		change := AttendanceBalanceChange{
			StudentID:     studentID,
			OldBillable:   false,
			NewBillable:   true,
			NewLessonDate: lessonOnPaidUntil,
		}

		next, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 5, 5, 0)
		assert.False(t, changed)
		assert.Equal(t, int64(5), next)
	})

	t.Run("Priority2_DateExpired_DeductsOneSession", func(t *testing.T) {
		lessonAfterPaidUntil := time.Date(2026, 11, 16, 18, 0, 0, 0, time.UTC)
		change := AttendanceBalanceChange{
			StudentID:     studentID,
			OldBillable:   false,
			NewBillable:   true,
			NewLessonDate: lessonAfterPaidUntil,
		}

		next, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 5, 5, 1)
		assert.True(t, changed)
		assert.Equal(t, int64(4), next)
	})

	t.Run("Priority2_NullPaidUntil_DeductsOneSession", func(t *testing.T) {
		lessonDate := time.Date(2026, 11, 10, 18, 0, 0, 0, time.UTC)
		change := AttendanceBalanceChange{
			StudentID:     studentID,
			OldBillable:   false,
			NewBillable:   true,
			NewLessonDate: lessonDate,
		}

		next, changed := ResolveAttendanceSessionBalance(change, nil, 1, 1, 1)
		assert.True(t, changed)
		assert.Equal(t, int64(0), next)
	})

	t.Run("Priority3_Debt_DateExpiredAndZeroSessions_StaysZero", func(t *testing.T) {
		lessonAfterPaidUntil := time.Date(2026, 11, 20, 18, 0, 0, 0, time.UTC)
		change := AttendanceBalanceChange{
			StudentID:     studentID,
			OldBillable:   false,
			NewBillable:   true,
			NewLessonDate: lessonAfterPaidUntil,
		}

		next, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 0, 0, 1)
		assert.False(t, changed)
		assert.Equal(t, int64(0), next)
	})

	t.Run("RefundOnDelete_SessionBackedVisit_RefundsOne", func(t *testing.T) {
		lessonDate := time.Date(2026, 11, 20, 18, 0, 0, 0, time.UTC)
		change := AttendanceBalanceChange{
			StudentID:     studentID,
			OldBillable:   true,
			OldLessonDate: lessonDate,
			NewBillable:   false,
		}

		// Bought 5 sessions, 1 remaining uncovered visit left after deleting the 2nd visit -> maxAllowed = 4
		next, changed := ResolveAttendanceSessionBalance(change, nil, 3, 5, 1)
		assert.True(t, changed)
		assert.Equal(t, int64(4), next)
	})

	t.Run("RefundOnDelete_DebtVisitWithNoSessionPacks_DoesNotFabricateSession", func(t *testing.T) {
		lessonDate := time.Date(2026, 11, 20, 18, 0, 0, 0, time.UTC)
		change := AttendanceBalanceChange{
			StudentID:     studentID,
			OldBillable:   true,
			OldLessonDate: lessonDate,
			NewBillable:   false,
		}

		next, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 0, 0, 0)
		assert.False(t, changed)
		assert.Equal(t, int64(0), next)
	})

	t.Run("RefundOnDelete_DebtVisitAfterExhaustedSessionPack_DoesNotFabricateSession", func(t *testing.T) {
		lessonDate := time.Date(2026, 11, 20, 18, 0, 0, 0, time.UTC)
		change := AttendanceBalanceChange{
			StudentID:     studentID,
			OldBillable:   true,
			OldLessonDate: lessonDate,
			NewBillable:   false,
		}

		// Bought 5 sessions, attended 6 lessons (6th was debt), deleted 6th -> 5 uncovered visits still remain!
		next, changed := ResolveAttendanceSessionBalance(change, nil, 0, 5, 5)
		assert.False(t, changed)
		assert.Equal(t, int64(0), next)
	})

	t.Run("LessonDateMovedFromCoveredToUncovered_DeductsSession", func(t *testing.T) {
		change := AttendanceBalanceChange{
			StudentID:     studentID,
			OldBillable:   true,
			OldLessonDate: time.Date(2026, 11, 10, 18, 0, 0, 0, time.UTC),
			NewBillable:   true,
			NewLessonDate: time.Date(2026, 11, 20, 18, 0, 0, 0, time.UTC),
		}

		next, changed := ResolveAttendanceSessionBalance(change, &paidUntil, 3, 5, 1)
		assert.True(t, changed)
		assert.Equal(t, int64(2), next)
	})
}
