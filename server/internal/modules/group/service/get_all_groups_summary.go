package service

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
	"golang.org/x/sync/errgroup"
)

func (s *Service) GetAllGroupsSummary(ctx context.Context, ownerID uuid.UUID) ([]domain.GroupSummary, error) {
	logger.InfoContext(ctx, "Getting all groups summary", "user_id", ownerID.String())
	groups, err := s.db.FindAllGroupsByOwnerID(ctx, ownerID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get all groups", "error", err, "user_id", ownerID.String())
		return nil, err
	}
	if len(groups) == 0 {
		logger.InfoContext(ctx, "No groups found for summary", "user_id", ownerID.String())
		return []domain.GroupSummary{}, nil
	}

	groupIDs := make([]uuid.UUID, 0, len(groups))
	for _, g := range groups {
		groupIDs = append(groupIDs, g.ID)
	}

	counts, err := s.db.CountStudentsByGroupIDs(ctx, db.CountStudentsByGroupIDsParams{
		OwnerID: ownerID,
		Column2: groupIDs,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count students by group IDs", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	countMap := make(map[uuid.UUID]int64, len(counts))
	for _, c := range counts {
		countMap[c.GroupID] = c.StudentCount
	}

	debtorCountMap := make(map[uuid.UUID]int64, len(groups))
	groupStudents, err := s.db.FindGroupStudentsByOwnerID(ctx, ownerID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch group students for debtor counts", "error", err, "user_id", ownerID.String())
		return nil, err
	}

	if len(groupStudents) > 0 {
		groupStudentIDsMap := make(map[uuid.UUID][]uuid.UUID)
		allStudentIDsMap := make(map[uuid.UUID]bool)
		for _, gs := range groupStudents {
			groupStudentIDsMap[gs.GroupID] = append(groupStudentIDsMap[gs.GroupID], gs.StudentID)
			allStudentIDsMap[gs.StudentID] = true
		}

		allStudentIDs := make([]uuid.UUID, 0, len(allStudentIDsMap))
		for sid := range allStudentIDsMap {
			allStudentIDs = append(allStudentIDs, sid)
		}

		var (
			stRows              []studentmodel.Student
			visitsByStudentID   map[uuid.UUID][]visitmodel.Visit
			balancesByStudentID map[uuid.UUID]txmodel.StudentBalance
		)

		g, gCtx := errgroup.WithContext(ctx)
		if s.studentProvider != nil {
			g.Go(func() error {
				rows, err := s.studentProvider.GetStudentsByIDs(gCtx, allStudentIDs, ownerID)
				if err != nil {
					return err
				}
				stRows = rows
				return nil
			})
		}
		if s.visitProvider != nil {
			g.Go(func() error {
				vm, err := s.visitProvider.GetVisitsByStudentIDs(gCtx, allStudentIDs, ownerID)
				if err != nil {
					return err
				}
				visitsByStudentID = vm
				return nil
			})
		}
		if s.transactionProvider != nil {
			g.Go(func() error {
				bm, err := s.transactionProvider.GetStudentBalancesByStudentIDs(gCtx, allStudentIDs, ownerID)
				if err != nil {
					return err
				}
				balancesByStudentID = bm
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			logger.ErrorContext(ctx, "Failed to fetch student dependencies for group debtor counts", "error", err, "user_id", ownerID.String())
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
				logger.WarnContext(ctx, "Failed to fetch lessons for debtor counts", "error", err, "user_id", ownerID.String())
			} else {
				for _, lr := range lRows {
					lessonsMap[lr.ID] = lr
				}
			}
		}

		today := time.Now().UTC()
		for _, grp := range groups {
			groupLessonsMap := make(map[uuid.UUID]lessonmodel.Lesson)
			for lid, l := range lessonsMap {
				if l.GroupID != nil && *l.GroupID == grp.ID {
					groupLessonsMap[lid] = l
				}
			}

			sids := groupStudentIDsMap[grp.ID]
			var countOfDebtors int64
			for _, sid := range sids {
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

				var statuses []studentmodel.StudentStatus
				if s.studentProvider != nil {
					statuses = s.studentProvider.ComputeStudentStatuses(st, stVisits, balancePtr, groupLessonsMap, groupLessonsMap, today)
				}
				for _, status := range statuses {
					if status == studentmodel.StudentStatusUnpaid {
						countOfDebtors++
						break
					}
				}
			}
			debtorCountMap[grp.ID] = countOfDebtors
		}
	}

	summaries := make([]domain.GroupSummary, 0, len(groups))
	for _, g := range groups {
		var defPrice *float64
		if g.DefaultPrice.Valid {
			f, _ := strconv.ParseFloat(g.DefaultPrice.String, 64)
			defPrice = &f
		}
		var note *string
		if g.Note.Valid {
			note = &g.Note.String
		}
		summaries = append(summaries, domain.GroupSummary{
			ID:              g.ID,
			Name:            g.Name,
			DefaultPrice:    defPrice,
			Note:            note,
			CreatedAt:       g.CreatedAt,
			CountOfStudents: countMap[g.ID],
			CountOfDebtors:  debtorCountMap[g.ID],
		})
	}

	logger.InfoContext(ctx, "Retrieved groups summary", "count", len(summaries), "user_id", ownerID.String())
	return summaries, nil
}
