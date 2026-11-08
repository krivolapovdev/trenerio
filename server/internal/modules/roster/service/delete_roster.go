package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func (s *Service) DeleteRoster(ctx context.Context, rosterID, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Deleting roster", "roster_id", rosterID.String(), "user_id", ownerID.String())
	_, err := s.db.FindRosterByIDAndOwnerID(ctx, db.FindRosterByIDAndOwnerIDParams{
		ID:      rosterID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Roster to delete not found", "roster_id", rosterID.String(), "user_id", ownerID.String())
			return pkgerrors.NewNotFound("Roster", rosterID, ownerID)
		}
		logger.ErrorContext(ctx, "Failed to find roster for deletion", "roster_id", rosterID.String(), "error", err, "user_id", ownerID.String())
		return err
	}

	if err := s.db.DeleteRosterByIDAndOwnerID(ctx, db.DeleteRosterByIDAndOwnerIDParams{
		ID:      rosterID,
		OwnerID: ownerID,
	}); err != nil {
		logger.ErrorContext(ctx, "Failed to delete roster record", "roster_id", rosterID.String(), "error", err, "user_id", ownerID.String())
		return err
	}

	logger.InfoContext(ctx, "Roster deleted successfully", "roster_id", rosterID.String(), "user_id", ownerID.String())
	return nil
}
