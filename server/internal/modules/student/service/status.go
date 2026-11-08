package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	lessonsvc "github.com/krivolapovdev/trenerio/internal/modules/lesson/service"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/model"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/model"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/model"
)

func ComputeStudentStatuses(
	student response.StudentResponse,
	visits []visitmodel.VisitResponse,
	balance *txmodel.StudentBalance,
	lessonsMap map[uuid.UUID]response.LessonResponse,
	today time.Time,
) []studentmodel.StudentStatus {
	return ComputeStudentStatusesWithVisitLessons(student, visits, balance, lessonsMap, lessonsMap, today)
}

func ComputeStudentStatusesWithVisitLessons(
	student response.StudentResponse,
	visits []visitmodel.VisitResponse,
	balance *txmodel.StudentBalance,
	lessonsMap map[uuid.UUID]response.LessonResponse,
	visitLessonsMap map[uuid.UUID]response.LessonResponse,
	_ time.Time,
) []studentmodel.StudentStatus {
	if len(visits) == 0 && !balance.HasPaymentActivity() {
		if student.Free {
			return []studentmodel.StudentStatus{studentmodel.StudentStatusFree}
		}
		return []studentmodel.StudentStatus{studentmodel.StudentStatusInactive}
	}

	statuses := make([]studentmodel.StudentStatus, 0)

	studentVisitLessons := make(map[uuid.UUID]response.LessonResponse)
	for _, v := range visits {
		if l, ok := visitLessonsMap[v.LessonID]; ok {
			studentVisitLessons[v.LessonID] = l
		}
	}

	lastVisitLesson := lessonsvc.FindLatestLesson(studentVisitLessons)
	if lastVisitLesson != nil {
		if v := visitmodel.FindVisitByLessonID(visits, lastVisitLesson.ID); v != nil {
			if v.Status.IsPresent() {
				statuses = append(statuses, studentmodel.StudentStatusPresent)
			} else {
				statuses = append(statuses, studentmodel.StudentStatusMissing)
			}
		}
	}

	if student.Free {
		statuses = append(statuses, studentmodel.StudentStatusFree)
		return statuses
	}

	studentLessons := make(map[uuid.UUID]response.LessonResponse)
	for _, v := range visits {
		if v.Type == visitmodel.VisitTypeFree || v.Status == visitmodel.VisitStatusExcused {
			continue
		}
		if l, ok := lessonsMap[v.LessonID]; ok {
			studentLessons[v.LessonID] = l
		}
	}

	lastLesson := lessonsvc.FindLatestLesson(studentLessons)
	if lastLesson == nil {
		statuses = append(statuses, studentmodel.StudentStatusPaid)
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
		if student.Free || v.Type == visitmodel.VisitTypeFree || v.Status == visitmodel.VisitStatusExcused {
			continue
		}
		if !v.Status.IsPresent() && v.Status != visitmodel.VisitStatusAbsent {
			continue
		}
		l, ok := lessonsMap[v.LessonID]
		if !ok {
			continue
		}
		if !txmodel.IsSubscriptionActive(paidUntil, l.Date) {
			uncoveredBillableVisits++
		}
	}

	if sessionsLeft > 0 || uncoveredBillableVisits <= totalSessionsAdded {
		statuses = append(statuses, studentmodel.StudentStatusPaid)
	} else {
		statuses = append(statuses, studentmodel.StudentStatusUnpaid)
	}

	return statuses
}
