package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/mapper"
)

func (s *Service) AddStudents(ctx context.Context, rosterID, ownerID uuid.UUID, studentIDs []uuid.UUID, note *string) (domain.Roster, error) {
	logger.InfoContext(ctx, "Adding students to roster", "roster_id", rosterID.String(), "count", len(studentIDs), "user_id", ownerID.String())
	roster, err := s.db.FindRosterByIDAndOwnerID(ctx, db.FindRosterByIDAndOwnerIDParams{
		ID:      rosterID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Roster{}, pkgerrors.NewNotFound("Roster", rosterID, ownerID)
		}
		return domain.Roster{}, err
	}

	uniqueStudentIDs := utils.UniqueUUIDs(studentIDs)
	if len(uniqueStudentIDs) == 0 {
		return s.GetRosterByID(ctx, rosterID, ownerID)
	}

	if err := s.validateStudentIDsExist(ctx, uniqueStudentIDs, ownerID); err != nil {
		return domain.Roster{}, err
	}

	var sqlNote sql.NullString
	if note != nil && strings.TrimSpace(*note) != "" {
		sqlNote = sql.NullString{String: strings.TrimSpace(*note), Valid: true}
	}

	err = s.withTx(ctx, func(qtx *db.Queries) error {
		for _, sid := range studentIDs {
			if _, err := qtx.CreateParticipant(ctx, db.CreateParticipantParams{
				RosterID:  rosterID,
				StudentID: sid,
				Note:      sqlNote,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to add students to roster", "roster_id", rosterID.String(), "error", err, "user_id", ownerID.String())
		return domain.Roster{}, err
	}

	students, err := s.getParticipantsForRoster(ctx, rosterID, ownerID)
	if err != nil {
		return domain.Roster{}, err
	}

	return mapper.ToRosterModel(roster, int64(len(students)), students), nil
}
