package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/student/mapper"
)

func (s *Service) CreateStudent(ctx context.Context, studentToCreate studentmodel.Student, ownerID uuid.UUID) (studentmodel.Student, error) {
	logger.InfoContext(ctx, "Creating student", "name", studentToCreate.FullName, "user_id", ownerID.String())
	fullName, err := validator.ValidateRequiredName(studentToCreate.FullName, "fullName")
	if err != nil {
		logger.WarnContext(ctx, "Student full name validation failed", "user_id", ownerID.String())
		return studentmodel.Student{}, err
	}

	var birthdate sql.NullTime
	if studentToCreate.Birthdate != nil {
		if err := validator.ValidateDateNotFuture(studentToCreate.Birthdate.Time, "birthdate"); err != nil {
			logger.WarnContext(ctx, "Student birthdate cannot be in the future", "birthdate", studentToCreate.Birthdate.Format(time.DateOnly), "user_id", ownerID.String())
			return studentmodel.Student{}, err
		}
		birthdate = sql.NullTime{Time: studentToCreate.Birthdate.Time, Valid: true}
	}

	taken, err := s.db.ExistsStudentByOwnerIDAndFullName(ctx, db.ExistsStudentByOwnerIDAndFullNameParams{
		OwnerID:  ownerID,
		FullName: fullName,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check student name uniqueness", "name", fullName, "error", err, "user_id", ownerID.String())
		return studentmodel.Student{}, err
	}
	if taken {
		logger.WarnContext(ctx, "Student with name already exists", "name", fullName, "user_id", ownerID.String())
		return studentmodel.Student{}, pkgerrors.NewConflict(
			ErrNameNotUnique,
			fmt.Sprintf("A student named \"%s\" already exists", fullName),
		)
	}
	var phone sql.NullString
	if studentToCreate.Phone != nil && *studentToCreate.Phone != "" {
		phone = sql.NullString{String: *studentToCreate.Phone, Valid: true}
	}
	var note sql.NullString
	if studentToCreate.Note != nil && *studentToCreate.Note != "" {
		note = sql.NullString{String: *studentToCreate.Note, Valid: true}
	}

	var created db.StudentsModuleStudent
	err = s.withTx(ctx, func(qtx *db.Queries) error {
		var err error
		created, err = qtx.CreateStudent(ctx, db.CreateStudentParams{
			OwnerID:   ownerID,
			FullName:  fullName,
			Birthdate: birthdate,
			Phone:     phone,
			Note:      note,
			Free:      studentToCreate.Free,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create student record in database", "name", fullName, "error", err, "user_id", ownerID.String())
			return err
		}

		return nil
	})
	if err != nil {
		return studentmodel.Student{}, err
	}

	logger.InfoContext(ctx, "Student created successfully", "student_id", created.ID.String(), "name", created.FullName, "user_id", ownerID.String())
	return mapper.ToStudentModel(created), nil
}
