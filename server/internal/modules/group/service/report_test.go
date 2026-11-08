package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/model"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGroupReport_DayGrid(t *testing.T) {
	groupID := uuid.New()
	annaID := uuid.New()
	ivanID := uuid.New()
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	firstLessonID := uuid.New()
	secondLessonID := uuid.New()

	lessons := []response.LessonResponse{
		{ID: uuid.New(), Date: time.Date(2026, 2, 3, 10, 0, 0, 0, time.UTC), CreatedAt: createdAt, GroupID: &groupID},
		{ID: firstLessonID, Date: time.Date(2026, 2, 5, 10, 0, 0, 0, time.UTC), CreatedAt: createdAt, GroupID: &groupID},
		{ID: secondLessonID, Date: time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC), CreatedAt: createdAt, GroupID: &groupID},
		{ID: uuid.New(), Date: time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC), CreatedAt: createdAt, GroupID: &groupID},
	}

	visits := map[uuid.UUID][]response.VisitResponse{
		annaID: {
			{ID: uuid.New(), LessonID: firstLessonID, StudentID: annaID, Status: visitmodel.VisitStatusPresent, CreatedAt: createdAt},
			{ID: uuid.New(), LessonID: secondLessonID, StudentID: annaID, Status: visitmodel.VisitStatusPresent, CreatedAt: createdAt},
		},
		ivanID: {
			{ID: uuid.New(), LessonID: firstLessonID, StudentID: ivanID, Status: visitmodel.VisitStatusPresent, CreatedAt: createdAt},
			{ID: uuid.New(), LessonID: secondLessonID, StudentID: ivanID, Status: visitmodel.VisitStatusAbsent, CreatedAt: createdAt},
		},
	}

	grp := response.GroupResponse{
		ID:        groupID,
		Name:      "Beginners",
		CreatedAt: createdAt,
	}

	students := []response.StudentResponse{
		{ID: annaID, FullName: "Anna Smirnova", CreatedAt: createdAt},
		{ID: ivanID, FullName: "Ivan Petrov", CreatedAt: createdAt},
	}

	report := BuildGroupReport(grp, 2026, 2, students, lessons, visits, map[uuid.UUID]txmodel.StudentBalance{}, today)

	assert.Equal(t, 28, report.DayCount)
	assert.Equal(t, []int{3, 5, 10}, report.LessonDays)
	require.Len(t, report.Students, 2)

	anna := report.Students[0]
	assert.Equal(t, "Anna Smirnova", anna.FullName)
	assert.Equal(t, []int{5, 10}, anna.PresentDays)
	assert.Equal(t, 2, anna.PresentCount)
	assert.Equal(t, 3, anna.LessonCount)

	ivan := report.Students[1]
	assert.Equal(t, "Ivan Petrov", ivan.FullName)
	assert.Equal(t, []int{5}, ivan.PresentDays)
	assert.Equal(t, 1, ivan.PresentCount)
	assert.Equal(t, 3, ivan.LessonCount)

	assert.Equal(t, 3, report.TotalPresent)
	assert.Equal(t, 6, report.TotalLessons)
}

func TestBuildGroupReport_MultipleLessonsInOneDay(t *testing.T) {
	groupID := uuid.New()
	annaID := uuid.New()
	ivanID := uuid.New()
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	morningLessonID := uuid.New()
	eveningLessonID := uuid.New()
	singleLessonID := uuid.New()

	lessons := []response.LessonResponse{
		{ID: eveningLessonID, Date: time.Date(2026, 2, 5, 18, 0, 0, 0, time.UTC), CreatedAt: createdAt, GroupID: &groupID},
		{ID: morningLessonID, Date: time.Date(2026, 2, 5, 10, 0, 0, 0, time.UTC), CreatedAt: createdAt, GroupID: &groupID},
		{ID: singleLessonID, Date: time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC), CreatedAt: createdAt, GroupID: &groupID},
	}

	visits := map[uuid.UUID][]response.VisitResponse{
		annaID: {
			{ID: uuid.New(), LessonID: morningLessonID, StudentID: annaID, Status: visitmodel.VisitStatusPresent, CreatedAt: createdAt},
			{ID: uuid.New(), LessonID: eveningLessonID, StudentID: annaID, Status: visitmodel.VisitStatusPresent, CreatedAt: createdAt},
			{ID: uuid.New(), LessonID: singleLessonID, StudentID: annaID, Status: visitmodel.VisitStatusAbsent, CreatedAt: createdAt},
		},
		ivanID: {
			{ID: uuid.New(), LessonID: morningLessonID, StudentID: ivanID, Status: visitmodel.VisitStatusPresent, CreatedAt: createdAt},
			{ID: uuid.New(), LessonID: eveningLessonID, StudentID: ivanID, Status: visitmodel.VisitStatusAbsent, CreatedAt: createdAt},
			{ID: uuid.New(), LessonID: singleLessonID, StudentID: ivanID, Status: visitmodel.VisitStatusPresent, CreatedAt: createdAt},
		},
	}

	grp := response.GroupResponse{
		ID:        groupID,
		Name:      "Beginners",
		CreatedAt: createdAt,
	}

	students := []response.StudentResponse{
		{ID: annaID, FullName: "Anna Smirnova", CreatedAt: createdAt},
		{ID: ivanID, FullName: "Ivan Petrov", CreatedAt: createdAt},
	}

	report := BuildGroupReport(grp, 2026, 2, students, lessons, visits, map[uuid.UUID]txmodel.StudentBalance{}, today)

	assert.Equal(t, 28, report.DayCount)
	assert.Equal(t, []int{5, 10}, report.LessonDays)
	require.Len(t, report.Students, 2)

	anna := report.Students[0]
	assert.Equal(t, "Anna Smirnova", anna.FullName)
	assert.Equal(t, []int{5, 5}, anna.PresentDays)
	assert.Equal(t, 2, anna.PresentCount)
	assert.Equal(t, 3, anna.LessonCount)
	require.Len(t, anna.DayMarks, 2)
	assert.Equal(t, 5, anna.DayMarks[0].Day)
	assert.Equal(t, []bool{true, true}, anna.DayMarks[0].Marks)
	assert.Equal(t, 10, anna.DayMarks[1].Day)
	assert.Equal(t, []bool{false}, anna.DayMarks[1].Marks)

	ivan := report.Students[1]
	assert.Equal(t, "Ivan Petrov", ivan.FullName)
	assert.Equal(t, []int{5, 10}, ivan.PresentDays)
	assert.Equal(t, 2, ivan.PresentCount)
	assert.Equal(t, 3, ivan.LessonCount)
	require.Len(t, ivan.DayMarks, 2)
	assert.Equal(t, 5, ivan.DayMarks[0].Day)
	assert.Equal(t, []bool{true, false}, ivan.DayMarks[0].Marks)
	assert.Equal(t, 10, ivan.DayMarks[1].Day)
	assert.Equal(t, []bool{true}, ivan.DayMarks[1].Marks)

	assert.Equal(t, 4, report.TotalPresent)
	assert.Equal(t, 6, report.TotalLessons)
}

func TestBuildGroupReport_OrderingAndEmpty(t *testing.T) {
	groupID := uuid.New()
	annaID := uuid.New()
	ivanID := uuid.New()
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	grp := response.GroupResponse{
		ID:        groupID,
		Name:      "Beginners",
		CreatedAt: createdAt,
	}

	lessons := []response.LessonResponse{
		{ID: uuid.New(), Date: time.Date(2026, 2, 3, 10, 0, 0, 0, time.UTC), CreatedAt: createdAt, GroupID: &groupID},
	}

	students := []response.StudentResponse{
		{ID: ivanID, FullName: "ivan petrov", CreatedAt: createdAt},
		{ID: annaID, FullName: "Anna Smirnova", CreatedAt: createdAt},
	}

	report := BuildGroupReport(grp, 2026, 2, students, lessons, map[uuid.UUID][]response.VisitResponse{}, map[uuid.UUID]txmodel.StudentBalance{}, today)

	require.Len(t, report.Students, 2)
	assert.Equal(t, "Anna Smirnova", report.Students[0].FullName)
	assert.Equal(t, "ivan petrov", report.Students[1].FullName)
	assert.Equal(t, 0, report.TotalPresent)
	assert.Equal(t, 2, report.TotalLessons)

	emptyReport := BuildGroupReport(grp, 2026, 2, []response.StudentResponse{}, []response.LessonResponse{}, map[uuid.UUID][]response.VisitResponse{}, map[uuid.UUID]txmodel.StudentBalance{}, today)
	assert.Empty(t, emptyReport.Students)
	assert.Equal(t, 0, emptyReport.TotalPresent)
	assert.Equal(t, 0, emptyReport.TotalLessons)
	assert.Equal(t, 28, emptyReport.DayCount)
}

func TestBuildGroupReport_PaidFlags(t *testing.T) {
	groupID := uuid.New()
	decID := uuid.New()
	noPayID := uuid.New()
	oct1ID := uuid.New()
	oct31ID := uuid.New()
	sessionPackID := uuid.New()
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	grp := response.GroupResponse{
		ID:        groupID,
		Name:      "Beginners",
		CreatedAt: createdAt,
	}

	lessons := []response.LessonResponse{
		{ID: uuid.New(), Date: time.Date(2026, 2, 3, 10, 0, 0, 0, time.UTC), CreatedAt: createdAt, GroupID: &groupID},
	}

	students := []response.StudentResponse{
		{ID: decID, FullName: "Anna ExpiresInDecember", CreatedAt: createdAt},
		{ID: noPayID, FullName: "Boris WithoutPayments", CreatedAt: createdAt},
		{ID: oct1ID, FullName: "Cyril ExpiresInOctober", CreatedAt: createdAt},
		{ID: oct31ID, FullName: "Daria PaidThroughOctober", CreatedAt: createdAt},
		{ID: sessionPackID, FullName: "Elena SessionPack", CreatedAt: createdAt},
	}

	dec31 := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	oct1 := time.Date(2026, 10, 1, 23, 59, 59, 0, time.UTC)
	oct31 := time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC)

	balances := map[uuid.UUID]txmodel.StudentBalance{
		decID:         {StudentID: decID, PaidUntil: &dec31, SessionsLeft: 0},
		oct1ID:        {StudentID: oct1ID, PaidUntil: &oct1, SessionsLeft: 0},
		oct31ID:       {StudentID: oct31ID, PaidUntil: &oct31, SessionsLeft: 0},
		sessionPackID: {StudentID: sessionPackID, PaidUntil: nil, SessionsLeft: 4},
	}

	reportFeb := BuildGroupReport(grp, 2026, 2, students, lessons, map[uuid.UUID][]response.VisitResponse{}, balances, today)
	var febPaid []bool
	for _, s := range reportFeb.Students {
		febPaid = append(febPaid, s.Paid)
	}
	assert.Equal(t, []bool{true, false, true, true, true}, febPaid)

	reportOct := BuildGroupReport(grp, 2026, 10, students, lessons, map[uuid.UUID][]response.VisitResponse{}, balances, today)
	var octPaid []bool
	for _, s := range reportOct.Students {
		octPaid = append(octPaid, s.Paid)
	}
	assert.Equal(t, []bool{true, false, false, true, true}, octPaid)

	reportDec := BuildGroupReport(grp, 2026, 12, students, lessons, map[uuid.UUID][]response.VisitResponse{}, balances, today)
	var decPaid []bool
	for _, s := range reportDec.Students {
		decPaid = append(decPaid, s.Paid)
	}
	assert.Equal(t, []bool{true, false, false, false, true}, decPaid)
}
