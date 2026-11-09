package http

import (
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

type LessonResponse = domain.Lesson

// @name LessonDetailsResponse
type LessonDetailsResponse = domain.LessonDetails

type StudentResponse = studentmodel.Student
type VisitResponse = visitmodel.Visit
