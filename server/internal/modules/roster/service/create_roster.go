package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/mapper"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

func (s *Service) CreateRoster(ctx context.Context, rosterToCreate domain.Roster, studentIDs []uuid.UUID, ownerID uuid.UUID) (domain.Roster, error) {
	logger.InfoContext(ctx, "Creating roster", "name", rosterToCreate.Name, "user_id", ownerID.String())
	name, err := validator.ValidateRequiredName(rosterToCreate.Name, "name")
	if err != nil {
		logger.WarnContext(ctx, "Roster name validation failed", "user_id", ownerID.String())
		return domain.Roster{}, err
	}

	var uniqueStudentIDs []uuid.UUID
	if len(studentIDs) > 0 {
		uniqueStudentIDs = utils.UniqueUUIDs(studentIDs)
		if err := s.validateStudentIDsExist(ctx, uniqueStudentIDs, ownerID); err != nil {
			return domain.Roster{}, err
		}
	}

	var created db.RostersModuleRoster
	err = s.withTx(ctx, func(qtx *db.Queries) error {
		var date sql.NullTime
		if rosterToCreate.Date != nil {
			date = sql.NullTime{Time: *rosterToCreate.Date, Valid: true}
		}

		var note sql.NullString
		if rosterToCreate.Note != nil && strings.TrimSpace(*rosterToCreate.Note) != "" {
			note = sql.NullString{String: strings.TrimSpace(*rosterToCreate.Note), Valid: true}
		}

		var err error
		created, err = qtx.CreateRoster(ctx, db.CreateRosterParams{
			OwnerID: ownerID,
			Name:    name,
			Date:    date,
			Note:    note,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create roster record", "name", name, "error", err, "user_id", ownerID.String())
			return err
		}

		for _, sid := range studentIDs {
			_, err := qtx.CreateParticipant(ctx, db.CreateParticipantParams{
				RosterID:  created.ID,
				StudentID: sid,
				Note:      sql.NullString{},
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to create participant", "roster_id", created.ID.String(), "student_id", sid.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create roster in database", "name", name, "error", err, "user_id", ownerID.String())
		return domain.Roster{}, err
	}

	var students []studentmodel.Student
	if len(uniqueStudentIDs) > 0 && s.studentProvider != nil {
		stList, err := s.studentProvider.GetStudentsByIDs(ctx, uniqueStudentIDs, ownerID)
		if err == nil {
			students = stList
		}
	}

	logger.InfoContext(ctx, "Roster created successfully", "roster_id", created.ID.String(), "name", created.Name, "user_id", ownerID.String())
	return mapper.ToRosterModel(created, int64(len(uniqueStudentIDs)), students), nil
}
