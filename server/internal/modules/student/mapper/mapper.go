package mapper

import (
	"github.com/krivolapovdev/trenerio/internal/core/types"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

func ToStudentModel(s db.StudentsModuleStudent) domain.Student {
	var birthdate *types.Date
	if s.Birthdate.Valid {
		birthdate = &types.Date{Time: s.Birthdate.Time}
	}
	var phone *string
	if s.Phone.Valid {
		phone = &s.Phone.String
	}
	var note *string
	if s.Note.Valid {
		note = &s.Note.String
	}
	return domain.Student{
		ID:        s.ID,
		FullName:  s.FullName,
		Birthdate: birthdate,
		Phone:     phone,
		Note:      note,
		CreatedAt: s.CreatedAt,
		Free:      s.Free,
	}
}

func ToStudentResponse(s db.StudentsModuleStudent) domain.Student {
	return ToStudentModel(s)
}
