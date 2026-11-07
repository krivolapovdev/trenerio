package student

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction"
	"github.com/krivolapovdev/trenerio/internal/modules/visit"
	"github.com/krivolapovdev/trenerio/internal/shared/response"
	"github.com/stretchr/testify/assert"
)

func TestStudentStatus_IndividualLessonLast(t *testing.T) {
	studentID := uuid.New()
	groupID := uuid.New()
	groupLessonID := uuid.New()
	individualLessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	st := response.StudentResponse{
		ID:        studentID,
		FullName:  "Test Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []response.VisitResponse{
		{ID: uuid.New(), LessonID: groupLessonID, StudentID: studentID, Status: visit.VisitStatusAbsent},
		{ID: uuid.New(), LessonID: individualLessonID, StudentID: studentID, Status: visit.VisitStatusPresent},
	}

	lessonsMap := map[uuid.UUID]response.LessonResponse{
		groupLessonID: {
			ID:        groupLessonID,
			GroupID:   &groupID,
			Date:      time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
			CreatedAt: createdAt,
		},
		individualLessonID: {
			ID:        individualLessonID,
			GroupID:   nil,
			Date:      time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC),
			CreatedAt: createdAt,
		},
	}

	statuses := ComputeStudentStatuses(st, visits, nil, lessonsMap, time.Now().UTC())

	assert.Contains(t, statuses, StudentStatusPresent)
	assert.NotContains(t, statuses, StudentStatusMissing)
}

func TestStudentStatus_GroupLessonLast(t *testing.T) {
	studentID := uuid.New()
	groupID := uuid.New()
	groupLessonID := uuid.New()
	individualLessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	st := response.StudentResponse{
		ID:        studentID,
		FullName:  "Test Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []response.VisitResponse{
		{ID: uuid.New(), LessonID: individualLessonID, StudentID: studentID, Status: visit.VisitStatusPresent},
		{ID: uuid.New(), LessonID: groupLessonID, StudentID: studentID, Status: visit.VisitStatusAbsent},
	}

	lessonsMap := map[uuid.UUID]response.LessonResponse{
		individualLessonID: {
			ID:        individualLessonID,
			GroupID:   nil,
			Date:      time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
			CreatedAt: createdAt,
		},
		groupLessonID: {
			ID:        groupLessonID,
			GroupID:   &groupID,
			Date:      time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC),
			CreatedAt: createdAt,
		},
	}

	statuses := ComputeStudentStatuses(st, visits, nil, lessonsMap, time.Now().UTC())

	assert.Contains(t, statuses, StudentStatusMissing)
	assert.NotContains(t, statuses, StudentStatusPresent)
}

func TestStudentStatus_PrefersLessonCreatedLastOnSameDay(t *testing.T) {
	studentID := uuid.New()
	groupID := uuid.New()
	groupLessonID := uuid.New()
	individualLessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	day := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)

	st := response.StudentResponse{
		ID:        studentID,
		FullName:  "Test Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []response.VisitResponse{
		{ID: uuid.New(), LessonID: groupLessonID, StudentID: studentID, Status: visit.VisitStatusAbsent},
		{ID: uuid.New(), LessonID: individualLessonID, StudentID: studentID, Status: visit.VisitStatusPresent},
	}

	lessonsMap := map[uuid.UUID]response.LessonResponse{
		groupLessonID: {
			ID:        groupLessonID,
			GroupID:   &groupID,
			Date:      day,
			CreatedAt: createdAt,
		},
		individualLessonID: {
			ID:        individualLessonID,
			GroupID:   nil,
			Date:      day,
			CreatedAt: createdAt.Add(2 * time.Hour),
		},
	}

	statuses := ComputeStudentStatuses(st, visits, nil, lessonsMap, time.Now().UTC())

	assert.Contains(t, statuses, StudentStatusPresent)
	assert.NotContains(t, statuses, StudentStatusMissing)
}

func TestStudentStatus_InactiveWhenNoVisitsAndNoBalanceActivity(t *testing.T) {
	studentID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	stNotFree := response.StudentResponse{
		ID:        studentID,
		FullName:  "Inactive Student",
		Free:      false,
		CreatedAt: createdAt,
	}
	statuses := ComputeStudentStatuses(stNotFree, nil, nil, nil, time.Now().UTC())
	assert.Equal(t, []StudentStatus{StudentStatusInactive}, statuses)

	zeroBalance := &transaction.StudentBalance{StudentID: studentID, PaidUntil: nil, SessionsLeft: 0}
	zeroStatuses := ComputeStudentStatuses(stNotFree, nil, zeroBalance, nil, time.Now().UTC())
	assert.Equal(t, []StudentStatus{StudentStatusInactive}, zeroStatuses)

	stFree := response.StudentResponse{
		ID:        studentID,
		FullName:  "Free Student",
		Free:      true,
		CreatedAt: createdAt,
	}
	freeStatuses := ComputeStudentStatuses(stFree, nil, nil, nil, time.Now().UTC())
	assert.Equal(t, []StudentStatus{StudentStatusFree}, freeStatuses)
}

func TestStudentStatus_PaidVsUnpaidAndSessionBalances(t *testing.T) {
	studentID := uuid.New()
	lessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	lastLessonDay := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)

	st := response.StudentResponse{
		ID:        studentID,
		FullName:  "Paying Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []response.VisitResponse{
		{ID: uuid.New(), LessonID: lessonID, StudentID: studentID, Status: visit.VisitStatusPresent},
	}

	lessonsMap := map[uuid.UUID]response.LessonResponse{
		lessonID: {
			ID:        lessonID,
			Date:      lastLessonDay,
			CreatedAt: createdAt,
		},
	}

	t.Run("ActiveSubscription_Paid", func(t *testing.T) {
		activeBal := &transaction.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &lastLessonDay,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, visits, activeBal, lessonsMap, time.Now().UTC())
		assert.Contains(t, statuses, StudentStatusPaid)
		assert.NotContains(t, statuses, StudentStatusUnpaid)
	})

	t.Run("ExpiredSubscription_ZeroSessions_Unpaid", func(t *testing.T) {
		expired := lastLessonDay.AddDate(0, 0, -1)
		expiredBal := &transaction.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &expired,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, visits, expiredBal, lessonsMap, time.Now().UTC())
		assert.Contains(t, statuses, StudentStatusUnpaid)
		assert.NotContains(t, statuses, StudentStatusPaid)
	})

	t.Run("ExpiredSubscription_WithReserveSessions_Paid", func(t *testing.T) {
		expired := lastLessonDay.AddDate(0, 0, -1)
		hybridBal := &transaction.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &expired,
			SessionsLeft: 3,
		}
		statuses := ComputeStudentStatuses(st, visits, hybridBal, lessonsMap, time.Now().UTC())
		assert.Contains(t, statuses, StudentStatusPaid)
		assert.NotContains(t, statuses, StudentStatusUnpaid)
	})

	t.Run("SessionPackageOnly_NoVisitsYet_PaidNotInactive", func(t *testing.T) {
		sessionBal := &transaction.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    nil,
			SessionsLeft: 8,
		}
		statuses := ComputeStudentStatuses(st, nil, sessionBal, nil, time.Now().UTC())
		assert.Equal(t, []StudentStatus{StudentStatusPaid}, statuses)
	})

	t.Run("SubscriptionWithPastPaidUntil_NoVisitsYet_PaidNotUnpaid", func(t *testing.T) {
		today := time.Date(2026, 11, 7, 15, 30, 0, 0, time.UTC)
		pastPaidUntil := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
		pastSubBal := &transaction.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &pastPaidUntil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, nil, pastSubBal, nil, today)
		assert.Equal(t, []StudentStatus{StudentStatusPaid}, statuses)
	})

	t.Run("SubscriptionWithFuturePaidUntil_NoVisitsYet_Paid", func(t *testing.T) {
		today := time.Date(2026, 11, 7, 15, 30, 0, 0, time.UTC)
		futurePaidUntil := time.Date(2026, 12, 7, 0, 0, 0, 0, time.UTC)
		futureSubBal := &transaction.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &futurePaidUntil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, nil, futureSubBal, nil, today)
		assert.Equal(t, []StudentStatus{StudentStatusPaid}, statuses)
	})

	t.Run("ExhaustedSessionPackage_WithVisits_Unpaid", func(t *testing.T) {
		exhaustedBal := &transaction.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    nil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, visits, exhaustedBal, lessonsMap, time.Now().UTC())
		assert.Equal(t, []StudentStatus{StudentStatusPresent, StudentStatusUnpaid}, statuses)
	})

	t.Run("SingleSessionPurchased_OneLessonAttended_Paid", func(t *testing.T) {
		bal := &transaction.StudentBalance{
			StudentID:          studentID,
			PaidUntil:          nil,
			SessionsLeft:       0,
			TotalSessionsAdded: 1,
		}
		statuses := ComputeStudentStatuses(st, visits, bal, lessonsMap, time.Now().UTC())
		assert.Equal(t, []StudentStatus{StudentStatusPresent, StudentStatusPaid}, statuses)
	})

	t.Run("SingleSessionPurchased_TwoLessonsAttended_Unpaid", func(t *testing.T) {
		lesson2ID := uuid.New()
		twoVisits := []response.VisitResponse{
			visits[0],
			{
				ID:        uuid.New(),
				LessonID:  lesson2ID,
				StudentID: studentID,
				Status:    visit.VisitStatusPresent,
				Type:      visit.VisitTypeRegular,
			},
		}
		twoLessonsMap := map[uuid.UUID]response.LessonResponse{
			lessonID: lessonsMap[lessonID],
			lesson2ID: {
				ID:        lesson2ID,
				Date:      lastLessonDay.AddDate(0, 0, 1),
				CreatedAt: createdAt,
			},
		}
		bal := &transaction.StudentBalance{
			StudentID:          studentID,
			PaidUntil:          nil,
			SessionsLeft:       0,
			TotalSessionsAdded: 1,
		}
		statuses := ComputeStudentStatuses(st, twoVisits, bal, twoLessonsMap, time.Now().UTC())
		assert.Equal(t, []StudentStatus{StudentStatusPresent, StudentStatusUnpaid}, statuses)
	})

	t.Run("FreeLessonOnly_NoBalance_PaidNotUnpaid", func(t *testing.T) {
		freeLessonID := uuid.New()
		freeLessonDay := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
		freeVisits := []response.VisitResponse{
			{
				ID:        uuid.New(),
				LessonID:  freeLessonID,
				StudentID: studentID,
				Status:    visit.VisitStatusPresent,
				Type:      visit.VisitTypeFree,
			},
		}
		freeLessonsMap := map[uuid.UUID]response.LessonResponse{
			freeLessonID: {
				ID:        freeLessonID,
				Date:      freeLessonDay,
				CreatedAt: createdAt,
			},
		}
		statuses := ComputeStudentStatuses(st, freeVisits, nil, freeLessonsMap, time.Now().UTC())
		assert.Equal(t, []StudentStatus{StudentStatusPresent, StudentStatusPaid}, statuses)
	})

	t.Run("BillableLessonPaid_FollowedByFreeLesson_RemainsPaid", func(t *testing.T) {
		regularLessonID := uuid.New()
		freeLessonID := uuid.New()
		regularDay := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
		freeDay := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
		paidUntil := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)

		twoVisits := []response.VisitResponse{
			{
				ID:        uuid.New(),
				LessonID:  regularLessonID,
				StudentID: studentID,
				Status:    visit.VisitStatusPresent,
				Type:      visit.VisitTypeRegular,
			},
			{
				ID:        uuid.New(),
				LessonID:  freeLessonID,
				StudentID: studentID,
				Status:    visit.VisitStatusPresent,
				Type:      visit.VisitTypeFree,
			},
		}
		mixedLessonsMap := map[uuid.UUID]response.LessonResponse{
			regularLessonID: {
				ID:        regularLessonID,
				Date:      regularDay,
				CreatedAt: createdAt,
			},
			freeLessonID: {
				ID:        freeLessonID,
				Date:      freeDay,
				CreatedAt: createdAt,
			},
		}
		bal := &transaction.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &paidUntil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, twoVisits, bal, mixedLessonsMap, time.Now().UTC())
		assert.Equal(t, []StudentStatus{StudentStatusPresent, StudentStatusPaid}, statuses)
	})

	t.Run("BillableLessonUnpaid_FollowedByFreeLesson_RemainsUnpaid", func(t *testing.T) {
		regularLessonID := uuid.New()
		freeLessonID := uuid.New()
		regularDay := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
		freeDay := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
		expiredPaidUntil := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

		twoVisits := []response.VisitResponse{
			{
				ID:        uuid.New(),
				LessonID:  regularLessonID,
				StudentID: studentID,
				Status:    visit.VisitStatusPresent,
				Type:      visit.VisitTypeRegular,
			},
			{
				ID:        uuid.New(),
				LessonID:  freeLessonID,
				StudentID: studentID,
				Status:    visit.VisitStatusPresent,
				Type:      visit.VisitTypeFree,
			},
		}
		mixedLessonsMap := map[uuid.UUID]response.LessonResponse{
			regularLessonID: {
				ID:        regularLessonID,
				Date:      regularDay,
				CreatedAt: createdAt,
			},
			freeLessonID: {
				ID:        freeLessonID,
				Date:      freeDay,
				CreatedAt: createdAt,
			},
		}
		bal := &transaction.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &expiredPaidUntil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, twoVisits, bal, mixedLessonsMap, time.Now().UTC())
		assert.Equal(t, []StudentStatus{StudentStatusPresent, StudentStatusUnpaid}, statuses)
	})
}

func TestStudentStatus_FreeStudent(t *testing.T) {
	studentID := uuid.New()
	lessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	lastLessonDay := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)

	st := response.StudentResponse{
		ID:        studentID,
		FullName:  "Free Student",
		Free:      true,
		CreatedAt: createdAt,
	}

	visits := []response.VisitResponse{
		{ID: uuid.New(), LessonID: lessonID, StudentID: studentID, Status: visit.VisitStatusPresent},
	}

	lessonsMap := map[uuid.UUID]response.LessonResponse{
		lessonID: {
			ID:        lessonID,
			Date:      lastLessonDay,
			CreatedAt: createdAt,
		},
	}

	statuses := ComputeStudentStatuses(st, visits, nil, lessonsMap, time.Now().UTC())
	assert.Contains(t, statuses, StudentStatusFree)
	assert.Contains(t, statuses, StudentStatusPresent)
	assert.NotContains(t, statuses, StudentStatusPaid)
	assert.NotContains(t, statuses, StudentStatusUnpaid)
}

func TestStudentStatus_GroupScopedVisits(t *testing.T) {
	studentID := uuid.New()
	groupID := uuid.New()
	groupLessonID := uuid.New()
	individualLessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	groupLessonDay := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	individualLessonDay := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)

	st := response.StudentResponse{
		ID:        studentID,
		FullName:  "Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []response.VisitResponse{
		{ID: uuid.New(), LessonID: groupLessonID, StudentID: studentID, Status: visit.VisitStatusAbsent},
		{ID: uuid.New(), LessonID: individualLessonID, StudentID: studentID, Status: visit.VisitStatusPresent},
	}

	everyLessonMap := map[uuid.UUID]response.LessonResponse{
		groupLessonID: {
			ID:        groupLessonID,
			GroupID:   &groupID,
			Date:      groupLessonDay,
			CreatedAt: createdAt,
		},
		individualLessonID: {
			ID:        individualLessonID,
			GroupID:   nil,
			Date:      individualLessonDay,
			CreatedAt: createdAt,
		},
	}

	groupLessonsMap := map[uuid.UUID]response.LessonResponse{
		groupLessonID: {
			ID:        groupLessonID,
			GroupID:   &groupID,
			Date:      groupLessonDay,
			CreatedAt: createdAt,
		},
	}

	paidUntil := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	balance := &transaction.StudentBalance{
		StudentID:    studentID,
		PaidUntil:    &paidUntil,
		SessionsLeft: 0,
	}

	statuses := ComputeStudentStatusesWithVisitLessons(st, visits, balance, everyLessonMap, groupLessonsMap, time.Now().UTC())
	assert.Contains(t, statuses, StudentStatusMissing)
	assert.Contains(t, statuses, StudentStatusUnpaid)
	assert.NotContains(t, statuses, StudentStatusPresent)
	assert.NotContains(t, statuses, StudentStatusPaid)
}

func TestStudentStatus_IgnoresLessonsForOtherStudents(t *testing.T) {
	student1ID := uuid.New()
	lesson1ID := uuid.New()
	lesson2ID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	st1 := response.StudentResponse{
		ID:        student1ID,
		FullName:  "Student 1",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits1 := []response.VisitResponse{
		{ID: uuid.New(), LessonID: lesson1ID, StudentID: student1ID, Status: visit.VisitStatusAbsent},
	}

	lessonsMap := map[uuid.UUID]response.LessonResponse{
		lesson1ID: {
			ID:        lesson1ID,
			Date:      time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC),
			CreatedAt: createdAt,
		},
		lesson2ID: {
			ID:        lesson2ID,
			Date:      time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC),
			CreatedAt: createdAt,
		},
	}

	paidUntil := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	balance1 := &transaction.StudentBalance{
		StudentID:    student1ID,
		PaidUntil:    &paidUntil,
		SessionsLeft: 0,
	}

	statuses := ComputeStudentStatuses(st1, visits1, balance1, lessonsMap, time.Now().UTC())

	assert.Contains(t, statuses, StudentStatusMissing)
	assert.NotContains(t, statuses, StudentStatusPresent)
	assert.Contains(t, statuses, StudentStatusPaid)
	assert.NotContains(t, statuses, StudentStatusUnpaid)
}
