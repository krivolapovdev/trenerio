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

func (s *Service) RemoveStudent(ctx context.Context, rosterID, studentID, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Removing student from roster", "roster_id", rosterID.String(), "student_id", studentID.String(), "user_id", ownerID.String())
	if _, err := s.db.FindRosterByIDAndOwnerID(ctx, db.FindRosterByIDAndOwnerIDParams{
		ID:      rosterID,
		OwnerID: ownerID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pkgerrors.NewNotFound("Roster", rosterID, ownerID)
		}
		return err
	}

	exists, err := s.db.ExistsParticipant(ctx, db.ExistsParticipantParams{
		RosterID:  rosterID,
		StudentID: studentID,
	})
	if err != nil {
		return err
	}
	if !exists {
		return pkgerrors.NewNotFound("Participant", studentID, ownerID)
	}

	return s.db.DeleteParticipantByStudentAndRosterID(ctx, db.DeleteParticipantByStudentAndRosterIDParams{
		RosterID:  rosterID,
		StudentID: studentID,
	})
}
