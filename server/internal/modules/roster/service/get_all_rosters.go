package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/mapper"
)

func (s *Service) GetAllRosters(ctx context.Context, ownerID uuid.UUID) ([]domain.Roster, error) {
	logger.InfoContext(ctx, "Getting all rosters", "user_id", ownerID.String())
	rosterRows, err := s.db.FindAllRostersByOwnerID(ctx, ownerID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get all rosters", "error", err, "user_id", ownerID.String())
		return nil, err
	}
	if len(rosterRows) == 0 {
		logger.InfoContext(ctx, "No rosters found", "user_id", ownerID.String())
		return []domain.Roster{}, nil
	}

	rosterIDs := make([]uuid.UUID, 0, len(rosterRows))
	for _, r := range rosterRows {
		rosterIDs = append(rosterIDs, r.ID)
	}

	counts, err := s.db.CountParticipantsByRosterIDs(ctx, db.CountParticipantsByRosterIDsParams{
		OwnerID: ownerID,
		Column2: rosterIDs,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count participants by roster IDs", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	countMap := make(map[uuid.UUID]int64, len(counts))
	for _, c := range counts {
		countMap[c.RosterID] = c.StudentCount
	}

	res := make([]domain.Roster, 0, len(rosterRows))
	for _, r := range rosterRows {
		res = append(res, mapper.ToRosterModel(r, countMap[r.ID], nil))
	}

	logger.InfoContext(ctx, "Retrieved all rosters", "count", len(res), "user_id", ownerID.String())
	return res, nil
}
