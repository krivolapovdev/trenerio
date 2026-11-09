package http

import (
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	studenthttp "github.com/krivolapovdev/trenerio/internal/modules/student/transport/http"
)

// @name RosterResponse
type RosterResponse = domain.Roster

// @name ParticipantResponse
type ParticipantResponse = domain.Participant

type StudentResponse = studenthttp.StudentResponse
