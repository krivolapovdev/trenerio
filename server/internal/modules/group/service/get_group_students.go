package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
	"golang.org/x/sync/errgroup"
)

func (s *Service) GetGroupStudents(ctx context.Context, groupID, ownerID uuid.UUID) ([]domain.GroupStudentSummary, error) {
	logger.InfoContext(ctx, "Getting students for group", "group_id", groupID.String(), "user_id", ownerID.String())
	if _, err := s.GetGroupByID(ctx, groupID, ownerID); err != nil {
		logger.WarnContext(ctx, "Group not found for getting students", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	studentLinks, err := s.db.FindGroupStudentsByGroupIDAndOwnerID(ctx, db.FindGroupStudentsByGroupIDAndOwnerIDParams{
		GroupID: groupID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch student links for group", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}
	if len(studentLinks) == 0 {
		logger.InfoContext(ctx, "No students enrolled in group", "group_id", groupID.String(), "user_id", ownerID.String())
		return []domain.GroupStudentSummary{}, nil
	}

	studentIDs := make([]uuid.UUID, 0, len(studentLinks))
	for _, sl := range studentLinks {
		studentIDs = append(studentIDs, sl.StudentID)
	}
	studentIDs = utils.UniqueUUIDs(studentIDs)

	var (
		stRows              []studentmodel.Student
		visitsByStudentID   map[uuid.UUID][]visitmodel.Visit
		balancesByStudentID map[uuid.UUID]txmodel.StudentBalance
	)

	g, gCtx := errgroup.WithContext(ctx)
	if s.studentProvider != nil {
		g.Go(func() error {
			rows, err := s.studentProvider.GetStudentsByIDs(gCtx, studentIDs, ownerID)
			if err != nil {
				return err
			}
			stRows = rows
			return nil
		})
	}
	if s.visitProvider != nil {
		g.Go(func() error {
			vm, err := s.visitProvider.GetVisitsByStudentIDs(gCtx, studentIDs, ownerID)
			if err != nil {
				return err
			}
			visitsByStudentID = vm
			return nil
		})
	}
	if s.transactionProvider != nil {
		g.Go(func() error {
			bm, err := s.transactionProvider.GetStudentBalancesByStudentIDs(gCtx, studentIDs, ownerID)
			if err != nil {
				return err
			}
			balancesByStudentID = bm
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logger.ErrorContext(ctx, "Failed to fetch group student dependencies concurrently", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	studentsMap := make(map[uuid.UUID]studentmodel.Student, len(stRows))
	for _, sr := range stRows {
		studentsMap[sr.ID] = sr
	}

	allLessonIDsMap := make(map[uuid.UUID]bool)
	for _, visits := range visitsByStudentID {
		for _, vr := range visits {
			allLessonIDsMap[vr.LessonID] = true
		}
	}

	allLessonIDs := make([]uuid.UUID, 0, len(allLessonIDsMap))
	for lid := range allLessonIDsMap {
		allLessonIDs = append(allLessonIDs, lid)
	}

	lessonsMap := make(map[uuid.UUID]lessonmodel.Lesson, len(allLessonIDs))
	if len(allLessonIDs) > 0 && s.lessonProvider != nil {
		lRows, err := s.lessonProvider.GetLessonsByIDs(ctx, allLessonIDs, ownerID)
		if err != nil {
			logger.WarnContext(ctx, "Failed to fetch lessons for group students", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		} else {
			for _, lr := range lRows {
				lessonsMap[lr.ID] = lr
			}
		}
	}

	logger.DebugContext(ctx, "Computing statuses for group students",
		"group_id", groupID.String(),
		"students_count", len(studentIDs),
		"lessons_count", len(lessonsMap),
		"user_id", ownerID.String(),
	)

	today := time.Now().UTC()
	groupStudents := make([]domain.GroupStudentSummary, 0, len(studentIDs))
	for _, sid := range studentIDs {
		st, ok := studentsMap[sid]
		if !ok {
			continue
		}
		stVisits := visitsByStudentID[sid]
		var balancePtr *txmodel.StudentBalance
		if bal, ok := balancesByStudentID[sid]; ok {
			b := bal
			balancePtr = &b
		}
		groupLessonsMap := make(map[uuid.UUID]lessonmodel.Lesson)
		for lid, l := range lessonsMap {
			if l.GroupID != nil && *l.GroupID == groupID {
				groupLessonsMap[lid] = l
			}
		}

		var statuses []studentmodel.StudentStatus
		if s.studentProvider != nil {
			statuses = s.studentProvider.ComputeStudentStatuses(st, stVisits, balancePtr, groupLessonsMap, groupLessonsMap, today)
		}

		groupStudents = append(groupStudents, domain.GroupStudentSummary{
			ID:        st.ID,
			FullName:  st.FullName,
			Birthdate: st.Birthdate,
			Phone:     phonePtr(st.Phone),
			Note:      st.Note,
			CreatedAt: st.CreatedAt,
			Free:      st.Free,
			Statuses:  statuses,
		})
	}

	logger.InfoContext(ctx, "Retrieved group students", "group_id", groupID.String(), "count", len(groupStudents), "user_id", ownerID.String())
	return groupStudents, nil
}

func phonePtr(p *string) *string {
	if p == nil || *p == "" {
		return nil
	}
	return p
}
