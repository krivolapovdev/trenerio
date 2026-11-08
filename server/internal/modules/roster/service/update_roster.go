package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/types"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/mapper"
)

func (s *Service) UpdateRoster(ctx context.Context, rosterID, ownerID uuid.UUID, updates map[string]any) (domain.Roster, error) {
	logger.InfoContext(ctx, "Updating roster", "roster_id", rosterID.String(), "user_id", ownerID.String())
	var updated db.RostersModuleRoster

	err := s.withTx(ctx, func(qtx *db.Queries) error {
		existing, err := qtx.FindRosterByIDAndOwnerID(ctx, db.FindRosterByIDAndOwnerIDParams{
			ID:      rosterID,
			OwnerID: ownerID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Roster to update not found", "roster_id", rosterID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Roster", rosterID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find roster for update", "roster_id", rosterID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		name := existing.Name
		if rawName, ok := updates["name"]; ok && rawName != nil {
			nameStr, isStr := rawName.(string)
			if !isStr {
				logger.WarnContext(ctx, "Updated roster name must not be blank", "roster_id", rosterID.String(), "user_id", ownerID.String())
				return pkgerrors.NewBadRequest("name must not be blank")
			}
			trimmed, err := validator.ValidateRequiredName(nameStr, "name")
			if err != nil {
				return err
			}
			name = trimmed
		}

		date := existing.Date
		if rawDate, ok := updates["date"]; ok {
			if rawDate == nil {
				date = sql.NullTime{Valid: false}
			} else {
				switch v := rawDate.(type) {
				case time.Time:
					date = sql.NullTime{Time: v, Valid: true}
				case string:
					t, err := time.Parse(time.RFC3339, v)
					if err != nil {
						logger.WarnContext(ctx, "Invalid date format for update", "roster_id", rosterID.String(), "user_id", ownerID.String())
						return pkgerrors.NewBadRequest("invalid date format, expected RFC3339")
					}
					date = sql.NullTime{Time: t, Valid: true}
				default:
					return pkgerrors.NewBadRequest("invalid date format")
				}
			}
		}

		note := existing.Note
		if rawNote, ok := updates["note"]; ok {
			if rawNote == nil {
				note = sql.NullString{Valid: false}
			} else if str, ok := rawNote.(string); ok {
				trimmed := strings.TrimSpace(str)
				if trimmed == "" {
					note = sql.NullString{Valid: false}
				} else {
					note = sql.NullString{String: trimmed, Valid: true}
				}
			}
		}

		if rawSIDs, ok := updates["studentIds"]; ok {
			requestedStudentIDs := types.ParseUUIDSlice(rawSIDs)
			if len(requestedStudentIDs) > 0 {
				if err := s.validateStudentIDsExist(ctx, requestedStudentIDs, ownerID); err != nil {
					return err
				}
			}

			currentLinks, err := qtx.FindParticipantsByRosterIDAndOwnerID(ctx, db.FindParticipantsByRosterIDAndOwnerIDParams{
				RosterID: rosterID,
				OwnerID:  ownerID,
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to find current participants for update", "roster_id", rosterID.String(), "error", err, "user_id", ownerID.String())
				return err
			}

			currentStudentIDs := make([]uuid.UUID, 0, len(currentLinks))
			for _, cl := range currentLinks {
				currentStudentIDs = append(currentStudentIDs, cl.StudentID)
			}

			toDelete, toCreate := utils.DiffUUIDs(currentStudentIDs, requestedStudentIDs)
			for _, sid := range toDelete {
				if err := qtx.DeleteParticipantByStudentAndRosterID(ctx, db.DeleteParticipantByStudentAndRosterIDParams{
					RosterID:  rosterID,
					StudentID: sid,
				}); err != nil {
					logger.ErrorContext(ctx, "Failed to unlink student during roster update", "roster_id", rosterID.String(), "student_id", sid.String(), "error", err, "user_id", ownerID.String())
					return err
				}
			}

			for _, sid := range toCreate {
				if _, err := qtx.CreateParticipant(ctx, db.CreateParticipantParams{
					RosterID:  rosterID,
					StudentID: sid,
					Note:      sql.NullString{},
				}); err != nil {
					logger.ErrorContext(ctx, "Failed to link student during roster update", "roster_id", rosterID.String(), "student_id", sid.String(), "error", err, "user_id", ownerID.String())
					return err
				}
			}
		}

		var errUpdate error
		updated, errUpdate = qtx.UpdateRoster(ctx, db.UpdateRosterParams{
			ID:      rosterID,
			OwnerID: ownerID,
			Name:    name,
			Date:    date,
			Note:    note,
		})
		if errUpdate != nil {
			logger.ErrorContext(ctx, "Failed to update roster record", "roster_id", rosterID.String(), "error", errUpdate, "user_id", ownerID.String())
			return errUpdate
		}

		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update roster in database", "roster_id", rosterID.String(), "error", err, "user_id", ownerID.String())
		return domain.Roster{}, err
	}

	students, err := s.getParticipantsForRoster(ctx, rosterID, ownerID)
	if err != nil {
		return domain.Roster{}, err
	}

	logger.InfoContext(ctx, "Roster updated successfully", "roster_id", updated.ID.String(), "name", updated.Name, "user_id", ownerID.String())
	return mapper.ToRosterModel(updated, int64(len(students)), students), nil
}
