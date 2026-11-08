package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	groupmodel "github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
	"golang.org/x/sync/errgroup"
)

func (s *Service) GetAllStudentsSummary(ctx context.Context, ownerID uuid.UUID) ([]StudentSummary, error) {
	logger.InfoContext(ctx, "Getting students overview", "user_id", ownerID.String())
	students, err := s.GetAllStudents(ctx, ownerID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get students for overview", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	if len(students) == 0 {
		logger.InfoContext(ctx, "No students found for overview", "user_id", ownerID.String())
		return []StudentSummary{}, nil
	}

	studentIDs := make([]uuid.UUID, 0, len(students))
	for _, st := range students {
		studentIDs = append(studentIDs, st.ID)
	}

	var (
		visitsMap        map[uuid.UUID][]visitmodel.Visit
		balancesMap      map[uuid.UUID]txmodel.StudentBalance
		studentGroupsMap map[uuid.UUID][]groupmodel.Group
	)

	g, gCtx := errgroup.WithContext(ctx)
	if s.visitProvider != nil {
		g.Go(func() error {
			vm, err := s.visitProvider.GetVisitsByStudentIDs(gCtx, studentIDs, ownerID)
			if err != nil {
				return err
			}
			visitsMap = vm
			return nil
		})
	}
	if s.transactionProvider != nil {
		g.Go(func() error {
			bm, err := s.transactionProvider.GetStudentBalancesByOwnerID(gCtx, ownerID)
			if err != nil {
				return err
			}
			balancesMap = bm
			return nil
		})
	}
	if s.groupProvider != nil {
		g.Go(func() error {
			gm, err := s.groupProvider.GetGroupsByStudentIDs(gCtx, studentIDs, ownerID)
			if err != nil {
				return err
			}
			studentGroupsMap = gm
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logger.ErrorContext(ctx, "Failed to fetch student overview dependencies", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	lessonIDsMap := make(map[uuid.UUID]bool)
	for _, vs := range visitsMap {
		for _, v := range vs {
			lessonIDsMap[v.LessonID] = true
		}
	}

	allLessonIDs := make([]uuid.UUID, 0, len(lessonIDsMap))
	for lid := range lessonIDsMap {
		allLessonIDs = append(allLessonIDs, lid)
	}

	lessonsMap := make(map[uuid.UUID]lessonmodel.Lesson, len(allLessonIDs))
	if len(allLessonIDs) > 0 && s.lessonProvider != nil {
		lRows, err := s.lessonProvider.GetLessonsByIDs(ctx, allLessonIDs, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch lessons for students overview", "error", err, "user_id", ownerID.String())
			return nil, err
		}
		for _, lr := range lRows {
			lessonsMap[lr.ID] = lr
		}
	}

	logger.DebugContext(ctx, "Computing student statuses for overview",
		"students_count", len(students),
		"lessons_count", len(lessonsMap),
		"user_id", ownerID.String(),
	)

	today := time.Now().UTC()
	summaries := make([]StudentSummary, 0, len(students))
	for _, st := range students {
		stVisits := visitsMap[st.ID]
		var balancePtr *txmodel.StudentBalance
		if bal, ok := balancesMap[st.ID]; ok {
			b := bal
			balancePtr = &b
		}

		statuses := ComputeStudentStatuses(st, stVisits, balancePtr, lessonsMap, today)
		stGroups := studentGroupsMap[st.ID]
		if stGroups == nil {
			stGroups = []groupmodel.Group{}
		}

		var paidUntil *time.Time
		var sessionsLeft int64
		if balancePtr != nil {
			paidUntil = balancePtr.PaidUntil
			sessionsLeft = balancePtr.SessionsLeft
		}

		summaries = append(summaries, StudentSummary{
			ID:            st.ID,
			FullName:      st.FullName,
			Birthdate:     st.Birthdate,
			Phone:         st.Phone,
			Note:          st.Note,
			CreatedAt:     st.CreatedAt,
			Free:          st.Free,
			Statuses:      statuses,
			StudentGroups: stGroups,
			PaidUntil:     paidUntil,
			SessionsLeft:  sessionsLeft,
		})
	}

	logger.InfoContext(ctx, "Students overview generated successfully", "count", len(summaries), "user_id", ownerID.String())
	return summaries, nil
}
