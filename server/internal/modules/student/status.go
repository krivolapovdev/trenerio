package student

import (
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction"
	"github.com/krivolapovdev/trenerio/internal/modules/visit"
	"github.com/krivolapovdev/trenerio/internal/shared/response"
)

func ComputeStudentStatuses(
	student response.StudentResponse,
	visits []visit.VisitResponse,
	balance *transaction.StudentBalance,
	lessonsMap map[uuid.UUID]response.LessonResponse,
	today time.Time,
) []StudentStatus {
	return ComputeStudentStatusesWithVisitLessons(student, visits, balance, lessonsMap, lessonsMap, today)
}

func ComputeStudentStatusesWithVisitLessons(
	student response.StudentResponse,
	visits []visit.VisitResponse,
	balance *transaction.StudentBalance,
	lessonsMap map[uuid.UUID]response.LessonResponse,
	visitLessonsMap map[uuid.UUID]response.LessonResponse,
	_ time.Time,
) []StudentStatus {
	if len(visits) == 0 && !balance.HasPaymentActivity() {
		if student.Free {
			return []StudentStatus{StudentStatusFree}
		}
		return []StudentStatus{StudentStatusInactive}
	}

	statuses := make([]StudentStatus, 0)

	studentVisitLessons := make(map[uuid.UUID]response.LessonResponse)
	for _, v := range visits {
		if l, ok := visitLessonsMap[v.LessonID]; ok {
			studentVisitLessons[v.LessonID] = l
		}
	}

	lastVisitLesson := lesson.FindLatestLesson(studentVisitLessons)
	if lastVisitLesson != nil {
		if v := visit.FindVisitByLessonID(visits, lastVisitLesson.ID); v != nil {
			if v.Status.IsPresent() {
				statuses = append(statuses, StudentStatusPresent)
			} else {
				statuses = append(statuses, StudentStatusMissing)
			}
		}
	}

	if student.Free {
		statuses = append(statuses, StudentStatusFree)
		return statuses
	}

	studentLessons := make(map[uuid.UUID]response.LessonResponse)
	for _, v := range visits {
		if v.Type == visit.VisitTypeFree || v.Status == visit.VisitStatusExcused {
			continue
		}
		if l, ok := lessonsMap[v.LessonID]; ok {
			studentLessons[v.LessonID] = l
		}
	}

	lastLesson := lesson.FindLatestLesson(studentLessons)
	if lastLesson == nil {
		statuses = append(statuses, StudentStatusPaid)
		return statuses
	}

	var paidUntil *time.Time
	var totalSessionsAdded int64
	var sessionsLeft int64
	if balance != nil {
		paidUntil = balance.PaidUntil
		totalSessionsAdded = balance.TotalSessionsAdded
		sessionsLeft = balance.SessionsLeft
	}

	var uncoveredBillableVisits int64
	for _, v := range visits {
		if student.Free || v.Type == visit.VisitTypeFree || v.Status == visit.VisitStatusExcused {
			continue
		}
		if !v.Status.IsPresent() && v.Status != visit.VisitStatusAbsent {
			continue
		}
		l, ok := lessonsMap[v.LessonID]
		if !ok {
			continue
		}
		if !transaction.IsSubscriptionActive(paidUntil, l.Date) {
			uncoveredBillableVisits++
		}
	}

	if sessionsLeft > 0 || uncoveredBillableVisits <= totalSessionsAdded {
		statuses = append(statuses, StudentStatusPaid)
	} else {
		statuses = append(statuses, StudentStatusUnpaid)
	}

	return statuses
}
