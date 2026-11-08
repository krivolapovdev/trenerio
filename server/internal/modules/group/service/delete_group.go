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

func (s *Service) DeleteGroup(ctx context.Context, groupID, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Deleting group", "group_id", groupID.String(), "user_id", ownerID.String())
	_, err := s.db.FindGroupByIDAndOwnerID(ctx, db.FindGroupByIDAndOwnerIDParams{
		ID:      groupID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Group to delete not found", "group_id", groupID.String(), "user_id", ownerID.String())
			return pkgerrors.NewNotFound("Group", groupID, ownerID)
		}
		logger.ErrorContext(ctx, "Failed to find group for deletion", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return err
	}

	if err := s.db.DeleteGroupByIDAndOwnerID(ctx, db.DeleteGroupByIDAndOwnerIDParams{
		ID:      groupID,
		OwnerID: ownerID,
	}); err != nil {
		logger.ErrorContext(ctx, "Failed to delete group record", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return err
	}

	logger.InfoContext(ctx, "Group deleted successfully", "group_id", groupID.String(), "user_id", ownerID.String())
	return nil
}
