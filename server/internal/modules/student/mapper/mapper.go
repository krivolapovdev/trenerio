package mapper

import (
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/core/types"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func ToStudentResponse(s db.StudentsModuleStudent) response.StudentResponse {
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
	return response.StudentResponse{
		ID:        s.ID,
		FullName:  s.FullName,
		Birthdate: birthdate,
		Phone:     phone,
		Note:      note,
		CreatedAt: s.CreatedAt,
		Free:      s.Free,
	}
}
