package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/group/mapper"
)

func (s *Service) CreateGroup(ctx context.Context, groupToCreate domain.Group, studentIDs []uuid.UUID, ownerID uuid.UUID) (domain.Group, error) {
	logger.InfoContext(ctx, "Creating group", "name", groupToCreate.Name, "user_id", ownerID.String())
	name, err := validator.ValidateRequiredName(groupToCreate.Name, "name")
	if err != nil {
		logger.WarnContext(ctx, "Group name validation failed", "user_id", ownerID.String())
		return domain.Group{}, err
	}

	if groupToCreate.DefaultPrice != nil && *groupToCreate.DefaultPrice <= 0 {
		logger.WarnContext(ctx, "Group default price must be greater than 0", "default_price", *groupToCreate.DefaultPrice, "user_id", ownerID.String())
		return domain.Group{}, pkgerrors.NewBadRequest("defaultPrice must be greater than 0")
	}

	taken, err := s.db.ExistsGroupByOwnerIDAndName(ctx, db.ExistsGroupByOwnerIDAndNameParams{
		OwnerID: ownerID,
		Name:    name,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check group name uniqueness", "name", name, "error", err, "user_id", ownerID.String())
		return domain.Group{}, err
	}
	if taken {
		logger.WarnContext(ctx, "Group with name already exists", "name", name, "user_id", ownerID.String())
		return domain.Group{}, pkgerrors.NewConflict(
			ErrGroupNameNotUnique,
			fmt.Sprintf("A group named \"%s\" already exists", name),
		)
	}

	var uniqueStudentIDs []uuid.UUID
	if len(studentIDs) > 0 && s.studentProvider != nil {
		uniqueStudentIDs = utils.UniqueUUIDs(studentIDs)
		if err := validateStudentIDsExist(ctx, s.studentProvider, uniqueStudentIDs, ownerID, "group creation"); err != nil {
			return domain.Group{}, err
		}
	}

	var created db.GroupsModuleGroup
	err = s.withTx(ctx, func(qtx *db.Queries) error {
		var defPrice sql.NullString
		if groupToCreate.DefaultPrice != nil {
			defPrice = sql.NullString{String: strconv.FormatFloat(*groupToCreate.DefaultPrice, 'f', 2, 64), Valid: true}
		}
		var note sql.NullString
		if groupToCreate.Note != nil && *groupToCreate.Note != "" {
			note = sql.NullString{String: *groupToCreate.Note, Valid: true}
		}

		var err error
		created, err = qtx.CreateGroup(ctx, db.CreateGroupParams{
			OwnerID:      ownerID,
			Name:         name,
			DefaultPrice: defPrice,
			Note:         note,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create group record", "name", name, "error", err, "user_id", ownerID.String())
			return err
		}

		for _, sid := range uniqueStudentIDs {
			_, err := qtx.CreateGroupStudent(ctx, db.CreateGroupStudentParams{
				OwnerID:   ownerID,
				GroupID:   created.ID,
				StudentID: sid,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to link student to group", "group_id", created.ID.String(), "student_id", sid.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create group in database", "name", name, "error", err, "user_id", ownerID.String())
		return domain.Group{}, err
	}

	logger.InfoContext(ctx, "Group created successfully", "group_id", created.ID.String(), "name", created.Name, "students_count", len(uniqueStudentIDs), "user_id", ownerID.String())
	return mapper.ToGroupResponse(created), nil
}
