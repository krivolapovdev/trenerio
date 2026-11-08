package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/group/mapper"
)

func (s *Service) GetGroupsByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID][]domain.Group, error) {
	if len(studentIDs) == 0 {
		logger.DebugContext(ctx, "No student IDs provided to get groups", "user_id", ownerID.String())
		return map[uuid.UUID][]domain.Group{}, nil
	}
	groupStudentRows, err := s.db.FindGroupStudentsByStudentIDsAndOwnerID(ctx, db.FindGroupStudentsByStudentIDsAndOwnerIDParams{
		OwnerID: ownerID,
		Column2: studentIDs,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch student group links", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	studentGroupIDsMap := make(map[uuid.UUID][]uuid.UUID)
	allGroupIDsMap := make(map[uuid.UUID]bool)
	for _, gsr := range groupStudentRows {
		studentGroupIDsMap[gsr.StudentID] = append(studentGroupIDsMap[gsr.StudentID], gsr.GroupID)
		allGroupIDsMap[gsr.GroupID] = true
	}

	allGroupIDs := make([]uuid.UUID, 0, len(allGroupIDsMap))
	for gid := range allGroupIDsMap {
		allGroupIDs = append(allGroupIDs, gid)
	}

	groupsMap := make(map[uuid.UUID]domain.Group, len(allGroupIDs))
	if len(allGroupIDs) > 0 {
		gRows, err := s.db.FindGroupsByIDsAndOwnerID(ctx, db.FindGroupsByIDsAndOwnerIDParams{
			OwnerID: ownerID,
			Column2: allGroupIDs,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch groups by IDs", "count", len(allGroupIDs), "error", err, "user_id", ownerID.String())
			return nil, err
		}
		for _, gr := range gRows {
			groupsMap[gr.ID] = mapper.ToGroupResponse(gr)
		}
	}

	groupsByStudent := make(map[uuid.UUID][]domain.Group, len(studentIDs))
	for _, sid := range studentIDs {
		gIDs := studentGroupIDsMap[sid]
		stGroups := make([]domain.Group, 0, len(gIDs))
		for _, gid := range gIDs {
			if grp, ok := groupsMap[gid]; ok {
				stGroups = append(stGroups, grp)
			}
		}
		groupsByStudent[sid] = stGroups
	}

	logger.DebugContext(ctx, "Retrieved groups by student IDs", "student_count", len(studentIDs), "group_count", len(groupsMap), "user_id", ownerID.String())
	return groupsByStudent, nil
}
