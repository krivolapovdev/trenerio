package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func (s *Service) SetStudentGroups(ctx context.Context, studentID, ownerID uuid.UUID, groupIDs []uuid.UUID) error {
	logger.InfoContext(ctx, "Setting student groups", "student_id", studentID.String(), "groups_count", len(groupIDs), "user_id", ownerID.String())
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		currentLinks, err := qtx.FindGroupStudentsByStudentIDAndOwnerID(ctx, db.FindGroupStudentsByStudentIDAndOwnerIDParams{
			StudentID: studentID,
			OwnerID:   ownerID,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to find current student groups", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		currentGroupIDs := make([]uuid.UUID, 0, len(currentLinks))
		for _, cl := range currentLinks {
			currentGroupIDs = append(currentGroupIDs, cl.GroupID)
		}

		toDelete, toCreate := utils.DiffUUIDs(currentGroupIDs, groupIDs)

		for _, gid := range toDelete {
			if err := qtx.DeleteGroupStudentByStudentAndGroupID(ctx, db.DeleteGroupStudentByStudentAndGroupIDParams{
				GroupID:   gid,
				StudentID: studentID,
				OwnerID:   ownerID,
			}); err != nil {
				logger.ErrorContext(ctx, "Failed to unlink student from group", "student_id", studentID.String(), "group_id", gid.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		for _, gid := range toCreate {
			_, err := qtx.CreateGroupStudent(ctx, db.CreateGroupStudentParams{
				OwnerID:   ownerID,
				GroupID:   gid,
				StudentID: studentID,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to link student to group", "student_id", studentID.String(), "group_id", gid.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		logger.DebugContext(ctx, "Student groups sync details", "student_id", studentID.String(), "unlinked", len(toDelete), "linked", len(toCreate), "user_id", ownerID.String())
		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to set student groups in database", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return err
	}
	logger.InfoContext(ctx, "Student groups set successfully", "student_id", studentID.String(), "groups_count", len(groupIDs), "user_id", ownerID.String())
	return nil
}
