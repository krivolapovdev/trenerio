package http

import (
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

// @name RosterResponse
type RosterResponse = domain.Roster

// @name ParticipantResponse
type ParticipantResponse = domain.Participant

type Student = studentmodel.Student
type StudentResponse = Student
