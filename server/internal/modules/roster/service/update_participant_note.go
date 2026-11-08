package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func (s *Service) UpdateParticipantNote(ctx context.Context, rosterID, studentID, ownerID uuid.UUID, note *string) error {
	logger.InfoContext(ctx, "Updating participant note", "roster_id", rosterID.String(), "student_id", studentID.String(), "user_id", ownerID.String())
	if _, err := s.db.FindRosterByIDAndOwnerID(ctx, db.FindRosterByIDAndOwnerIDParams{
		ID:      rosterID,
		OwnerID: ownerID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pkgerrors.NewNotFound("Roster", rosterID, ownerID)
		}
		return err
	}

	var sqlNote sql.NullString
	if note != nil && strings.TrimSpace(*note) != "" {
		sqlNote = sql.NullString{String: strings.TrimSpace(*note), Valid: true}
	}

	_, err := s.db.UpdateParticipantNote(ctx, db.UpdateParticipantNoteParams{
		RosterID:  rosterID,
		StudentID: studentID,
		Note:      sqlNote,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pkgerrors.NewNotFound("Participant", studentID, ownerID)
		}
		logger.ErrorContext(ctx, "Failed to update participant note", "roster_id", rosterID.String(), "student_id", studentID.String(), "error", err)
		return err
	}

	return nil
}
