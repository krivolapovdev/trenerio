package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/types"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/group/mapper"
)

func (s *Service) UpdateGroup(ctx context.Context, groupID, ownerID uuid.UUID, updates map[string]any) (domain.Group, error) {
	logger.InfoContext(ctx, "Updating group", "group_id", groupID.String(), "user_id", ownerID.String())
	var updated db.GroupsModuleGroup

	err := s.withTx(ctx, func(qtx *db.Queries) error {
		existing, err := qtx.FindGroupByIDAndOwnerID(ctx, db.FindGroupByIDAndOwnerIDParams{
			ID:      groupID,
			OwnerID: ownerID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Group to update not found", "group_id", groupID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Group", groupID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find group for update", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		name := existing.Name
		if rawName, ok := updates["name"]; ok && rawName != nil {
			nameStr, isStr := rawName.(string)
			if !isStr {
				logger.WarnContext(ctx, "Updated group name must not be blank", "group_id", groupID.String(), "user_id", ownerID.String())
				return pkgerrors.NewBadRequest("name must not be blank")
			}
			trimmed, err := validator.ValidateRequiredName(nameStr, "name")
			if err != nil {
				logger.WarnContext(ctx, "Updated group name validation failed", "group_id", groupID.String(), "user_id", ownerID.String())
				return err
			}
			taken, err := qtx.ExistsGroupByOwnerIDAndNameAndIDNot(ctx, db.ExistsGroupByOwnerIDAndNameAndIDNotParams{
				OwnerID: ownerID,
				Name:    trimmed,
				ID:      groupID,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to check group name uniqueness for update", "name", trimmed, "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
			if taken {
				logger.WarnContext(ctx, "Group name already exists for update", "name", trimmed, "group_id", groupID.String(), "user_id", ownerID.String())
				return pkgerrors.NewConflict(
					ErrGroupNameNotUnique,
					fmt.Sprintf("A group named \"%s\" already exists", trimmed),
				)
			}
			name = trimmed
		}

		defaultPrice := existing.DefaultPrice
		if rawDP, ok := updates["defaultPrice"]; ok {
			if rawDP == nil {
				defaultPrice = sql.NullString{Valid: false}
			} else {
				var priceVal float64
				switch v := rawDP.(type) {
				case float64:
					priceVal = v
				case int:
					priceVal = float64(v)
				case int64:
					priceVal = float64(v)
				case string:
					p, err := strconv.ParseFloat(v, 64)
					if err != nil {
						logger.WarnContext(ctx, "Invalid defaultPrice format for update", "group_id", groupID.String(), "user_id", ownerID.String())
						return pkgerrors.NewBadRequest("invalid defaultPrice format")
					}
					priceVal = p
				default:
					logger.WarnContext(ctx, "Invalid defaultPrice type for update", "group_id", groupID.String(), "user_id", ownerID.String())
					return pkgerrors.NewBadRequest("invalid defaultPrice format")
				}
				if priceVal <= 0 {
					logger.WarnContext(ctx, "Group default price must be greater than 0 for update", "group_id", groupID.String(), "user_id", ownerID.String())
					return pkgerrors.NewBadRequest("defaultPrice must be greater than 0")
				}
				defaultPrice = sql.NullString{String: strconv.FormatFloat(priceVal, 'f', 2, 64), Valid: true}
			}
		}

		note := existing.Note
		if rawNote, ok := updates["note"]; ok {
			if rawNote == nil {
				note = sql.NullString{Valid: false}
			} else if str, ok := rawNote.(string); ok {
				note = sql.NullString{String: str, Valid: true}
			}
		}

		if rawSIDs, ok := updates["studentIds"]; ok && s.studentProvider != nil {
			requestedStudentIDs := types.ParseUUIDSlice(rawSIDs)
			if len(requestedStudentIDs) > 0 {
				if err := validateStudentIDsExist(ctx, s.studentProvider, requestedStudentIDs, ownerID, "group update"); err != nil {
					return err
				}
			}

			currentLinks, err := qtx.FindGroupStudentsByGroupIDAndOwnerID(ctx, db.FindGroupStudentsByGroupIDAndOwnerIDParams{
				GroupID: groupID,
				OwnerID: ownerID,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to find current group students for update", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
				return err
			}

			currentStudentIDs := make([]uuid.UUID, 0, len(currentLinks))
			for _, cl := range currentLinks {
				currentStudentIDs = append(currentStudentIDs, cl.StudentID)
			}

			toDelete, toCreate := utils.DiffUUIDs(currentStudentIDs, requestedStudentIDs)
			deletedCount := 0
			for _, sid := range toDelete {
				if err := qtx.DeleteGroupStudentByStudentAndGroupID(ctx, db.DeleteGroupStudentByStudentAndGroupIDParams{
					GroupID:   groupID,
					StudentID: sid,
					OwnerID:   ownerID,
				}); err != nil {
					logger.ErrorContext(ctx, "Failed to unlink student during group update", "group_id", groupID.String(), "student_id", sid.String(), "error", err, "user_id", ownerID.String())
					return err
				}
				deletedCount++
			}

			createdCount := 0
			for _, sid := range toCreate {
				_, err := qtx.CreateGroupStudent(ctx, db.CreateGroupStudentParams{
					OwnerID:   ownerID,
					GroupID:   groupID,
					StudentID: sid,
				})
				if err != nil {
					logger.ErrorContext(ctx, "Failed to link student during group update", "group_id", groupID.String(), "student_id", sid.String(), "error", err, "user_id", ownerID.String())
					return err
				}
				createdCount++
			}

			logger.DebugContext(ctx, "Group students synced during update", "group_id", groupID.String(), "unlinked", deletedCount, "linked", createdCount, "user_id", ownerID.String())
		}

		updated, err = qtx.UpdateGroup(ctx, db.UpdateGroupParams{
			ID:           groupID,
			OwnerID:      ownerID,
			Name:         name,
			DefaultPrice: defaultPrice,
			Note:         note,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to update group record", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
			return err
		}
		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update group in database", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return domain.Group{}, err
	}

	logger.InfoContext(ctx, "Group updated successfully", "group_id", updated.ID.String(), "name", updated.Name, "user_id", ownerID.String())
	return mapper.ToGroupResponse(updated), nil
}
