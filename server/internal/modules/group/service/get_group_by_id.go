package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/group/mapper"
)

func (s *Service) GetGroupByID(ctx context.Context, groupID, ownerID uuid.UUID) (domain.Group, error) {
	logger.InfoContext(ctx, "Getting group by id", "group_id", groupID.String(), "user_id", ownerID.String())
	g, err := s.db.FindGroupByIDAndOwnerID(ctx, db.FindGroupByIDAndOwnerIDParams{
		ID:      groupID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Group not found", "group_id", groupID.String(), "user_id", ownerID.String())
			return domain.Group{}, pkgerrors.NewNotFound("Group", groupID, ownerID)
		}
		logger.ErrorContext(ctx, "Failed to get group by id", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return domain.Group{}, err
	}
	logger.DebugContext(ctx, "Retrieved group by id", "group_id", groupID.String(), "user_id", ownerID.String())
	return mapper.ToGroupResponse(g), nil
}
