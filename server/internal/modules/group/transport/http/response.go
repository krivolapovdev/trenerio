package http

import (
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	lessonhttp "github.com/krivolapovdev/trenerio/internal/modules/lesson/transport/http"
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

type LessonResponse = lessonhttp.LessonResponse
