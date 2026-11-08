package mapper

import (
	"time"

	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

func ToRosterModel(r db.RostersModuleRoster, countOfStudents int64, students []studentmodel.Student) domain.Roster {
	var date *time.Time
	if r.Date.Valid {
		date = &r.Date.Time
	}

	var note *string
	if r.Note.Valid {
		note = &r.Note.String
	}

	if students == nil {
		students = []studentmodel.Student{}
	}

	return domain.Roster{
		ID:              r.ID,
		Name:            r.Name,
		Date:            date,
		Note:            note,
		CreatedAt:       r.CreatedAt,
		CountOfStudents: countOfStudents,
		Students:        students,
	}
}
