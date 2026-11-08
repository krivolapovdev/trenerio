package http

import (
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

// @name LessonResponse
type LessonResponse = domain.Lesson

// @name LessonDetailsResponse
type LessonDetailsResponse = domain.LessonDetails

// @name StudentResponse
type StudentResponse = studentmodel.Student

// @name VisitResponse
type VisitResponse = visitmodel.Visit
