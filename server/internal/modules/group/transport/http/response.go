package http

import (
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
)

type GroupResponse = domain.Group

// @name GroupStudentResponse
type GroupStudentResponse = domain.GroupStudent

// @name GroupSummaryResponse
type GroupSummaryResponse = domain.GroupSummary

// @name GroupStudentSummaryResponse
type GroupStudentSummaryResponse = domain.GroupStudentSummary

// @name GroupReportDayMarkResponse
type GroupReportDayMarkResponse = domain.GroupReportDayMark

type GroupReportStudentResponse = domain.GroupReportStudent

// @name GroupReportResponse
type GroupReportResponse = domain.GroupReport

type Lesson = lessonmodel.Lesson
