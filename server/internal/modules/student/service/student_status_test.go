package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
	"github.com/stretchr/testify/assert"
)

func TestStudentStatus_IndividualLessonLast(t *testing.T) {
	studentID := uuid.New()
	groupID := uuid.New()
	groupLessonID := uuid.New()
	individualLessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	st := studentmodel.Student{
		ID:        studentID,
		FullName:  "Test Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []visitmodel.Visit{
		{ID: uuid.New(), LessonID: groupLessonID, StudentID: studentID, Status: visitmodel.VisitStatusAbsent},
		{ID: uuid.New(), LessonID: individualLessonID, StudentID: studentID, Status: visitmodel.VisitStatusPresent},
	}

	lessonsMap := map[uuid.UUID]lessonmodel.Lesson{
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

	assert.Contains(t, statuses, studentmodel.StudentStatusPresent)
	assert.NotContains(t, statuses, studentmodel.StudentStatusMissing)
}

func TestStudentStatus_GroupLessonLast(t *testing.T) {
	studentID := uuid.New()
	groupID := uuid.New()
	groupLessonID := uuid.New()
	individualLessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	st := studentmodel.Student{
		ID:        studentID,
		FullName:  "Test Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []visitmodel.Visit{
		{ID: uuid.New(), LessonID: individualLessonID, StudentID: studentID, Status: visitmodel.VisitStatusPresent},
		{ID: uuid.New(), LessonID: groupLessonID, StudentID: studentID, Status: visitmodel.VisitStatusAbsent},
	}

	lessonsMap := map[uuid.UUID]lessonmodel.Lesson{
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

	assert.Contains(t, statuses, studentmodel.StudentStatusMissing)
	assert.NotContains(t, statuses, studentmodel.StudentStatusPresent)
}

func TestStudentStatus_PrefersLessonCreatedLastOnSameDay(t *testing.T) {
	studentID := uuid.New()
	groupID := uuid.New()
	groupLessonID := uuid.New()
	individualLessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	day := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)

	st := studentmodel.Student{
		ID:        studentID,
		FullName:  "Test Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []visitmodel.Visit{
		{ID: uuid.New(), LessonID: groupLessonID, StudentID: studentID, Status: visitmodel.VisitStatusAbsent},
		{ID: uuid.New(), LessonID: individualLessonID, StudentID: studentID, Status: visitmodel.VisitStatusPresent},
	}

	lessonsMap := map[uuid.UUID]lessonmodel.Lesson{
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

	assert.Contains(t, statuses, studentmodel.StudentStatusPresent)
	assert.NotContains(t, statuses, studentmodel.StudentStatusMissing)
}

func TestStudentStatus_InactiveWhenNoVisitsAndNoBalanceActivity(t *testing.T) {
	studentID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	stNotFree := studentmodel.Student{
		ID:        studentID,
		FullName:  "Inactive Student",
		Free:      false,
		CreatedAt: createdAt,
	}
	statuses := ComputeStudentStatuses(stNotFree, nil, nil, nil, time.Now().UTC())
	assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusInactive}, statuses)

	zeroBalance := &txmodel.StudentBalance{StudentID: studentID, PaidUntil: nil, SessionsLeft: 0}
	zeroStatuses := ComputeStudentStatuses(stNotFree, nil, zeroBalance, nil, time.Now().UTC())
	assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusInactive}, zeroStatuses)

	stFree := studentmodel.Student{
		ID:        studentID,
		FullName:  "Free Student",
		Free:      true,
		CreatedAt: createdAt,
	}
	freeStatuses := ComputeStudentStatuses(stFree, nil, nil, nil, time.Now().UTC())
	assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusFree}, freeStatuses)
}

func TestStudentStatus_PaidVsUnpaidAndSessionBalances(t *testing.T) {
	studentID := uuid.New()
	lessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	lastLessonDay := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)

	st := studentmodel.Student{
		ID:        studentID,
		FullName:  "Paying Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []visitmodel.Visit{
		{ID: uuid.New(), LessonID: lessonID, StudentID: studentID, Status: visitmodel.VisitStatusPresent},
	}

	lessonsMap := map[uuid.UUID]lessonmodel.Lesson{
		lessonID: {
			ID:        lessonID,
			Date:      lastLessonDay,
			CreatedAt: createdAt,
		},
	}

	t.Run("ActiveSubscription_Paid", func(t *testing.T) {
		activeBal := &txmodel.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &lastLessonDay,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, visits, activeBal, lessonsMap, time.Now().UTC())
		assert.Contains(t, statuses, studentmodel.StudentStatusPaid)
		assert.NotContains(t, statuses, studentmodel.StudentStatusUnpaid)
	})

	t.Run("ExpiredSubscription_ZeroSessions_Unpaid", func(t *testing.T) {
		expired := lastLessonDay.AddDate(0, 0, -1)
		expiredBal := &txmodel.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &expired,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, visits, expiredBal, lessonsMap, time.Now().UTC())
		assert.Contains(t, statuses, studentmodel.StudentStatusUnpaid)
		assert.NotContains(t, statuses, studentmodel.StudentStatusPaid)
	})

	t.Run("ExpiredSubscription_WithReserveSessions_Paid", func(t *testing.T) {
		expired := lastLessonDay.AddDate(0, 0, -1)
		hybridBal := &txmodel.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &expired,
			SessionsLeft: 3,
		}
		statuses := ComputeStudentStatuses(st, visits, hybridBal, lessonsMap, time.Now().UTC())
		assert.Contains(t, statuses, studentmodel.StudentStatusPaid)
		assert.NotContains(t, statuses, studentmodel.StudentStatusUnpaid)
	})

	t.Run("SessionPackageOnly_NoVisitsYet_PaidNotInactive", func(t *testing.T) {
		sessionBal := &txmodel.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    nil,
			SessionsLeft: 8,
		}
		statuses := ComputeStudentStatuses(st, nil, sessionBal, nil, time.Now().UTC())
		assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusPaid}, statuses)
	})

	t.Run("SubscriptionWithPastPaidUntil_NoVisitsYet_PaidNotUnpaid", func(t *testing.T) {
		today := time.Date(2026, 11, 7, 15, 30, 0, 0, time.UTC)
		pastPaidUntil := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
		pastSubBal := &txmodel.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &pastPaidUntil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, nil, pastSubBal, nil, today)
		assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusPaid}, statuses)
	})

	t.Run("SubscriptionWithFuturePaidUntil_NoVisitsYet_Paid", func(t *testing.T) {
		today := time.Date(2026, 11, 7, 15, 30, 0, 0, time.UTC)
		futurePaidUntil := time.Date(2026, 12, 7, 0, 0, 0, 0, time.UTC)
		futureSubBal := &txmodel.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &futurePaidUntil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, nil, futureSubBal, nil, today)
		assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusPaid}, statuses)
	})

	t.Run("ExhaustedSessionPackage_WithVisits_Unpaid", func(t *testing.T) {
		exhaustedBal := &txmodel.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    nil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, visits, exhaustedBal, lessonsMap, time.Now().UTC())
		assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusPresent, studentmodel.StudentStatusUnpaid}, statuses)
	})

	t.Run("SingleSessionPurchased_OneLessonAttended_Paid", func(t *testing.T) {
		bal := &txmodel.StudentBalance{
			StudentID:          studentID,
			PaidUntil:          nil,
			SessionsLeft:       0,
			TotalSessionsAdded: 1,
		}
		statuses := ComputeStudentStatuses(st, visits, bal, lessonsMap, time.Now().UTC())
		assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusPresent, studentmodel.StudentStatusPaid}, statuses)
	})

	t.Run("SingleSessionPurchased_TwoLessonsAttended_Unpaid", func(t *testing.T) {
		lesson2ID := uuid.New()
		twoVisits := []visitmodel.Visit{
			visits[0],
			{
				ID:        uuid.New(),
				LessonID:  lesson2ID,
				StudentID: studentID,
				Status:    visitmodel.VisitStatusPresent,
				Type:      visitmodel.VisitTypeRegular,
			},
		}
		twoLessonsMap := map[uuid.UUID]lessonmodel.Lesson{
			lessonID: lessonsMap[lessonID],
			lesson2ID: {
				ID:        lesson2ID,
				Date:      lastLessonDay.AddDate(0, 0, 1),
				CreatedAt: createdAt,
			},
		}
		bal := &txmodel.StudentBalance{
			StudentID:          studentID,
			PaidUntil:          nil,
			SessionsLeft:       0,
			TotalSessionsAdded: 1,
		}
		statuses := ComputeStudentStatuses(st, twoVisits, bal, twoLessonsMap, time.Now().UTC())
		assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusPresent, studentmodel.StudentStatusUnpaid}, statuses)
	})

	t.Run("FreeLessonOnly_NoBalance_PaidNotUnpaid", func(t *testing.T) {
		freeLessonID := uuid.New()
		freeLessonDay := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
		freeVisits := []visitmodel.Visit{
			{
				ID:        uuid.New(),
				LessonID:  freeLessonID,
				StudentID: studentID,
				Status:    visitmodel.VisitStatusPresent,
				Type:      visitmodel.VisitTypeFree,
			},
		}
		freeLessonsMap := map[uuid.UUID]lessonmodel.Lesson{
			freeLessonID: {
				ID:        freeLessonID,
				Date:      freeLessonDay,
				CreatedAt: createdAt,
			},
		}
		statuses := ComputeStudentStatuses(st, freeVisits, nil, freeLessonsMap, time.Now().UTC())
		assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusPresent, studentmodel.StudentStatusPaid}, statuses)
	})

	t.Run("BillableLessonPaid_FollowedByFreeLesson_RemainsPaid", func(t *testing.T) {
		regularLessonID := uuid.New()
		freeLessonID := uuid.New()
		regularDay := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
		freeDay := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
		paidUntil := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)

		twoVisits := []visitmodel.Visit{
			{
				ID:        uuid.New(),
				LessonID:  regularLessonID,
				StudentID: studentID,
				Status:    visitmodel.VisitStatusPresent,
				Type:      visitmodel.VisitTypeRegular,
			},
			{
				ID:        uuid.New(),
				LessonID:  freeLessonID,
				StudentID: studentID,
				Status:    visitmodel.VisitStatusPresent,
				Type:      visitmodel.VisitTypeFree,
			},
		}
		mixedLessonsMap := map[uuid.UUID]lessonmodel.Lesson{
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
		bal := &txmodel.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &paidUntil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, twoVisits, bal, mixedLessonsMap, time.Now().UTC())
		assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusPresent, studentmodel.StudentStatusPaid}, statuses)
	})

	t.Run("BillableLessonUnpaid_FollowedByFreeLesson_RemainsUnpaid", func(t *testing.T) {
		regularLessonID := uuid.New()
		freeLessonID := uuid.New()
		regularDay := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
		freeDay := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
		expiredPaidUntil := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

		twoVisits := []visitmodel.Visit{
			{
				ID:        uuid.New(),
				LessonID:  regularLessonID,
				StudentID: studentID,
				Status:    visitmodel.VisitStatusPresent,
				Type:      visitmodel.VisitTypeRegular,
			},
			{
				ID:        uuid.New(),
				LessonID:  freeLessonID,
				StudentID: studentID,
				Status:    visitmodel.VisitStatusPresent,
				Type:      visitmodel.VisitTypeFree,
			},
		}
		mixedLessonsMap := map[uuid.UUID]lessonmodel.Lesson{
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
		bal := &txmodel.StudentBalance{
			StudentID:    studentID,
			PaidUntil:    &expiredPaidUntil,
			SessionsLeft: 0,
		}
		statuses := ComputeStudentStatuses(st, twoVisits, bal, mixedLessonsMap, time.Now().UTC())
		assert.Equal(t, []studentmodel.StudentStatus{studentmodel.StudentStatusPresent, studentmodel.StudentStatusUnpaid}, statuses)
	})
}

func TestStudentStatus_FreeStudent(t *testing.T) {
	studentID := uuid.New()
	lessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	lastLessonDay := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)

	st := studentmodel.Student{
		ID:        studentID,
		FullName:  "Free Student",
		Free:      true,
		CreatedAt: createdAt,
	}

	visits := []visitmodel.Visit{
		{ID: uuid.New(), LessonID: lessonID, StudentID: studentID, Status: visitmodel.VisitStatusPresent},
	}

	lessonsMap := map[uuid.UUID]lessonmodel.Lesson{
		lessonID: {
			ID:        lessonID,
			Date:      lastLessonDay,
			CreatedAt: createdAt,
		},
	}

	statuses := ComputeStudentStatuses(st, visits, nil, lessonsMap, time.Now().UTC())
	assert.Contains(t, statuses, studentmodel.StudentStatusFree)
	assert.Contains(t, statuses, studentmodel.StudentStatusPresent)
	assert.NotContains(t, statuses, studentmodel.StudentStatusPaid)
	assert.NotContains(t, statuses, studentmodel.StudentStatusUnpaid)
}

func TestStudentStatus_GroupScopedVisits(t *testing.T) {
	studentID := uuid.New()
	groupID := uuid.New()
	groupLessonID := uuid.New()
	individualLessonID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	groupLessonDay := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	individualLessonDay := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)

	st := studentmodel.Student{
		ID:        studentID,
		FullName:  "Student",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits := []visitmodel.Visit{
		{ID: uuid.New(), LessonID: groupLessonID, StudentID: studentID, Status: visitmodel.VisitStatusAbsent},
		{ID: uuid.New(), LessonID: individualLessonID, StudentID: studentID, Status: visitmodel.VisitStatusPresent},
	}

	everyLessonMap := map[uuid.UUID]lessonmodel.Lesson{
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

	groupLessonsMap := map[uuid.UUID]lessonmodel.Lesson{
		groupLessonID: {
			ID:        groupLessonID,
			GroupID:   &groupID,
			Date:      groupLessonDay,
			CreatedAt: createdAt,
		},
	}

	paidUntil := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	balance := &txmodel.StudentBalance{
		StudentID:    studentID,
		PaidUntil:    &paidUntil,
		SessionsLeft: 0,
	}

	statuses := ComputeStudentStatusesWithVisitLessons(st, visits, balance, everyLessonMap, groupLessonsMap, time.Now().UTC())
	assert.Contains(t, statuses, studentmodel.StudentStatusMissing)
	assert.Contains(t, statuses, studentmodel.StudentStatusUnpaid)
	assert.NotContains(t, statuses, studentmodel.StudentStatusPresent)
	assert.NotContains(t, statuses, studentmodel.StudentStatusPaid)
}

func TestStudentStatus_IgnoresLessonsForOtherStudents(t *testing.T) {
	student1ID := uuid.New()
	lesson1ID := uuid.New()
	lesson2ID := uuid.New()
	createdAt := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	st1 := studentmodel.Student{
		ID:        student1ID,
		FullName:  "Student 1",
		Free:      false,
		CreatedAt: createdAt,
	}

	visits1 := []visitmodel.Visit{
		{ID: uuid.New(), LessonID: lesson1ID, StudentID: student1ID, Status: visitmodel.VisitStatusAbsent},
	}

	lessonsMap := map[uuid.UUID]lessonmodel.Lesson{
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
	balance1 := &txmodel.StudentBalance{
		StudentID:    student1ID,
		PaidUntil:    &paidUntil,
		SessionsLeft: 0,
	}

	statuses := ComputeStudentStatuses(st1, visits1, balance1, lessonsMap, time.Now().UTC())

	assert.Contains(t, statuses, studentmodel.StudentStatusMissing)
	assert.NotContains(t, statuses, studentmodel.StudentStatusPresent)
	assert.Contains(t, statuses, studentmodel.StudentStatusPaid)
	assert.NotContains(t, statuses, studentmodel.StudentStatusUnpaid)
}
