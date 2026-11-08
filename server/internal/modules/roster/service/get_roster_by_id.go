package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/mapper"
)

func (s *Service) GetRosterByID(ctx context.Context, rosterID, ownerID uuid.UUID) (domain.Roster, error) {
	logger.InfoContext(ctx, "Getting roster by id", "roster_id", rosterID.String(), "user_id", ownerID.String())
	r, err := s.db.FindRosterByIDAndOwnerID(ctx, db.FindRosterByIDAndOwnerIDParams{
		ID:      rosterID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Roster not found", "roster_id", rosterID.String(), "user_id", ownerID.String())
			return domain.Roster{}, pkgerrors.NewNotFound("Roster", rosterID, ownerID)
		}
		logger.ErrorContext(ctx, "Failed to get roster by id", "roster_id", rosterID.String(), "error", err, "user_id", ownerID.String())
		return domain.Roster{}, err
	}

	students, err := s.getParticipantsForRoster(ctx, rosterID, ownerID)
	if err != nil {
		return domain.Roster{}, err
	}

	return mapper.ToRosterModel(r, int64(len(students)), students), nil
}
