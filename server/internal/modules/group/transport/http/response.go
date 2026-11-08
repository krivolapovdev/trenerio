package http

import (
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
)

// @name GroupResponse
type GroupResponse = domain.Group

// @name GroupStudentResponse
type GroupStudentResponse = domain.GroupStudent

// @name GroupSummaryResponse
type GroupSummaryResponse = domain.GroupSummary

// @name GroupStudentSummaryResponse
type GroupStudentSummaryResponse = domain.GroupStudentSummary

// @name GroupReportDayMarkResponse
type GroupReportDayMarkResponse = domain.GroupReportDayMark

// @name GroupReportStudentResponse
type GroupReportStudentResponse = domain.GroupReportStudent

// @name GroupReportResponse
type GroupReportResponse = domain.GroupReport

// @name LessonResponse
type LessonResponse = lessonmodel.Lesson
