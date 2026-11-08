package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

func (s *Service) GetRosterStudents(ctx context.Context, rosterID, ownerID uuid.UUID) ([]studentmodel.Student, error) {
	logger.InfoContext(ctx, "Getting students for roster", "roster_id", rosterID.String(), "user_id", ownerID.String())
	if _, err := s.db.FindRosterByIDAndOwnerID(ctx, db.FindRosterByIDAndOwnerIDParams{
		ID:      rosterID,
		OwnerID: ownerID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Roster not found", "roster_id", rosterID.String(), "user_id", ownerID.String())
			return nil, pkgerrors.NewNotFound("Roster", rosterID, ownerID)
		}
		return nil, err
	}

	return s.getParticipantsForRoster(ctx, rosterID, ownerID)
}
