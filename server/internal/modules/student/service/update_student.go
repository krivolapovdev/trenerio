package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/types"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/student/mapper"
)

func (s *Service) UpdateStudent(ctx context.Context, studentID, ownerID uuid.UUID, updates map[string]any) (studentmodel.Student, error) {
	logger.InfoContext(ctx, "Updating student", "student_id", studentID.String(), "user_id", ownerID.String())
	var updated db.StudentsModuleStudent

	err := s.withTx(ctx, func(qtx *db.Queries) error {
		existing, err := qtx.FindStudentByIDAndOwnerID(ctx, db.FindStudentByIDAndOwnerIDParams{
			ID:      studentID,
			OwnerID: ownerID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Student to update not found", "student_id", studentID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Student", studentID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find student for update", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		fullName := existing.FullName
		if rawName, ok := updates["fullName"]; ok && rawName != nil {
			nameStr, isStr := rawName.(string)
			if !isStr {
				logger.WarnContext(ctx, "Updated student name must not be blank", "student_id", studentID.String(), "user_id", ownerID.String())
				return pkgerrors.NewBadRequest("fullName must not be blank")
			}
			trimmed, err := validator.ValidateRequiredName(nameStr, "fullName")
			if err != nil {
				logger.WarnContext(ctx, "Updated student name validation failed", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
			taken, err := qtx.ExistsStudentByOwnerIDAndFullNameAndIDNot(ctx, db.ExistsStudentByOwnerIDAndFullNameAndIDNotParams{
				OwnerID:  ownerID,
				FullName: trimmed,
				ID:       studentID,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to check student name uniqueness for update", "name", trimmed, "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
			if taken {
				logger.WarnContext(ctx, "Student name already exists for update", "name", trimmed, "student_id", studentID.String(), "user_id", ownerID.String())
				return pkgerrors.NewConflict(
					ErrNameNotUnique,
					fmt.Sprintf("A student named \"%s\" already exists", trimmed),
				)
			}
			fullName = trimmed
		}

		birthdate := existing.Birthdate
		if rawBD, ok := updates["birthdate"]; ok {
			if rawBD == nil {
				birthdate = sql.NullTime{Valid: false}
			} else if s, ok := rawBD.(string); ok && s != "" {
				t, err := time.Parse(time.DateOnly, s)
				if err != nil {
					t, err = time.Parse(time.RFC3339, s)
				}
				if err == nil {
					if err := validator.ValidateDateNotFuture(t, "birthdate"); err != nil {
						logger.WarnContext(ctx, "Updated student birthdate cannot be in the future", "birthdate", s, "user_id", ownerID.String())
						return err
					}
					birthdate = sql.NullTime{Time: t, Valid: true}
				}
			}
		}

		phone := existing.Phone
		if rawPhone, ok := updates["phone"]; ok {
			if rawPhone == nil {
				phone = sql.NullString{Valid: false}
			} else if s, ok := rawPhone.(string); ok {
				phone = sql.NullString{String: s, Valid: true}
			}
		}

		note := existing.Note
		if rawNote, ok := updates["note"]; ok {
			if rawNote == nil {
				note = sql.NullString{Valid: false}
			} else if s, ok := rawNote.(string); ok {
				note = sql.NullString{String: s, Valid: true}
			}
		}

		free := existing.Free
		if rawFree, ok := updates["free"]; ok && rawFree != nil {
			if b, ok := rawFree.(bool); ok {
				free = b
			}
		}

		groupKey := ""
		if _, ok := updates["groupIds"]; ok {
			groupKey = "groupIds"
		} else if _, ok := updates["groupId"]; ok {
			groupKey = "groupId"
		}

		if groupKey != "" && s.groupProvider != nil {
			requestedGroupIDs := types.ParseUUIDSlice(updates[groupKey])
			if err := s.groupProvider.SetStudentGroups(ctx, studentID, ownerID, requestedGroupIDs); err != nil {
				logger.ErrorContext(ctx, "Failed to set student groupIds", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		updated, err = qtx.UpdateStudent(ctx, db.UpdateStudentParams{
			ID:        studentID,
			OwnerID:   ownerID,
			FullName:  fullName,
			Birthdate: birthdate,
			Phone:     phone,
			Note:      note,
			Free:      free,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to update student record", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return err
		}
		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update student transaction", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return studentmodel.Student{}, err
	}

	logger.InfoContext(ctx, "Student updated successfully", "student_id", updated.ID.String(), "name", updated.FullName, "user_id", ownerID.String())
	return mapper.ToStudentModel(updated), nil
}
