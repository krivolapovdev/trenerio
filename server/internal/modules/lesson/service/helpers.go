package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

func (s *Service) fetchStudentFreeFlagsTx(ctx context.Context, qtx *db.Queries, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID]bool, error) {
	freeMap := make(map[uuid.UUID]bool, len(studentIDs))
	if len(studentIDs) == 0 {
		return freeMap, nil
	}

	if qtx != nil {
		rows, err := qtx.FindStudentsByIDsAndOwnerID(ctx, db.FindStudentsByIDsAndOwnerIDParams{
			OwnerID: ownerID,
			Column2: studentIDs,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch student free flags for attendance balance sync", "error", err, "user_id", ownerID.String())
			return nil, err
		}
		for _, r := range rows {
			freeMap[r.ID] = r.Free
		}
		return freeMap, nil
	}

	if s.studentProvider != nil {
		students, err := s.studentProvider.GetStudentsByIDs(ctx, studentIDs, ownerID)
		if err != nil {
			return nil, err
		}
		for _, st := range students {
			freeMap[st.ID] = st.Free
		}
	}
	return freeMap, nil
}

func deduplicateStudentVisits(visits []visitmodel.StudentVisit) []visitmodel.StudentVisit {
	if len(visits) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]int, len(visits))
	unique := make([]visitmodel.StudentVisit, 0, len(visits))
	for _, sv := range visits {
		sv.Normalize()
		if idx, exists := seen[sv.StudentID]; exists {
			unique[idx] = sv
		} else {
			seen[sv.StudentID] = len(unique)
			unique = append(unique, sv)
		}
	}
	return unique
}
