package service

import (
	"context"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

func (s *Service) validateStudentIDsExist(ctx context.Context, ids []uuid.UUID, ownerID uuid.UUID) error {
	if s.studentProvider == nil || len(ids) == 0 {
		return nil
	}
	found, err := s.studentProvider.GetStudentsByIDs(ctx, ids, ownerID)
	if err != nil {
		return err
	}
	if len(found) != len(ids) {
		return pkgerrors.NewBadRequest("One or more student IDs not found")
	}
	return nil
}

func (s *Service) getParticipantsForRoster(ctx context.Context, rosterID, ownerID uuid.UUID) ([]studentmodel.Student, error) {
	participantRows, err := s.db.FindParticipantsByRosterIDAndOwnerID(ctx, db.FindParticipantsByRosterIDAndOwnerIDParams{
		RosterID: rosterID,
		OwnerID:  ownerID,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get participants for roster", "roster_id", rosterID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}
	if len(participantRows) == 0 {
		return []studentmodel.Student{}, nil
	}

	studentIDs := make([]uuid.UUID, 0, len(participantRows))
	for _, p := range participantRows {
		studentIDs = append(studentIDs, p.StudentID)
	}

	if s.studentProvider != nil {
		stList, err := s.studentProvider.GetStudentsByIDs(ctx, studentIDs, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch student details for roster", "roster_id", rosterID.String(), "error", err, "user_id", ownerID.String())
			return nil, err
		}
		studentMap := make(map[uuid.UUID]studentmodel.Student, len(stList))
		for _, st := range stList {
			studentMap[st.ID] = st
		}

		result := make([]studentmodel.Student, 0, len(participantRows))
		for _, p := range participantRows {
			if st, ok := studentMap[p.StudentID]; ok {
				item := st
				if p.Note.Valid && p.Note.String != "" {
					val := p.Note.String
					item.Note = &val
				} else {
					item.Note = nil
				}
				result = append(result, item)
			}
		}
		return result, nil
	}

	return []studentmodel.Student{}, nil
}
