package service

import (
	"context"
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

func (s *Service) GetGroupReport(ctx context.Context, groupID, ownerID uuid.UUID, year, month int) (domain.GroupReport, error) {
	logger.InfoContext(ctx, "Getting group report", "group_id", groupID.String(), "year", year, "month", month, "user_id", ownerID.String())
	grp, err := s.GetGroupByID(ctx, groupID, ownerID)
	if err != nil {
		logger.WarnContext(ctx, "Group not found for report", "group_id", groupID.String(), "user_id", ownerID.String())
		return domain.GroupReport{}, err
	}

	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := time.Date(year, time.Month(month)+1, 0, 23, 59, 59, 999999999, time.UTC)

	var (
		lessons      []lessonmodel.Lesson
		studentLinks []db.GroupsModuleGroupStudent
	)

	g, gCtx := errgroup.WithContext(ctx)
	if s.lessonProvider != nil {
		g.Go(func() error {
			lList, err := s.lessonProvider.GetLessonsByGroupIDAndDateRange(gCtx, groupID, ownerID, firstDay, lastDay)
			if err != nil {
				return err
			}
			lessons = lList
			return nil
		})
	}
	g.Go(func() error {
		sl, err := s.db.FindGroupStudentsByGroupIDAndOwnerID(gCtx, db.FindGroupStudentsByGroupIDAndOwnerIDParams{
			GroupID: groupID,
			OwnerID: ownerID,
		})
		if err != nil {
			logger.WarnContext(gCtx, "Failed to fetch group student links for report", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		} else {
			studentLinks = sl
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.ErrorContext(ctx, "Failed to fetch group report base dependencies", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return domain.GroupReport{}, err
	}

	lessonIDs := make([]uuid.UUID, 0, len(lessons))
	for _, lr := range lessons {
		lessonIDs = append(lessonIDs, lr.ID)
	}

	var lessonVisits []visitmodel.Visit
	if len(lessonIDs) > 0 && s.visitProvider != nil {
		lv, err := s.visitProvider.GetVisitsByLessonIDs(ctx, lessonIDs, ownerID)
		if err != nil {
			logger.WarnContext(ctx, "Failed to fetch visits for group report lessons", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		} else {
			lessonVisits = lv
		}
	}

	studentIDSet := make(map[uuid.UUID]bool)
	studentIDOrder := make([]uuid.UUID, 0)
	for _, sl := range studentLinks {
		if !studentIDSet[sl.StudentID] {
			studentIDSet[sl.StudentID] = true
			studentIDOrder = append(studentIDOrder, sl.StudentID)
		}
	}
	for _, lv := range lessonVisits {
		if !studentIDSet[lv.StudentID] {
			studentIDSet[lv.StudentID] = true
			studentIDOrder = append(studentIDOrder, lv.StudentID)
		}
	}

	visitsByStudentID := make(map[uuid.UUID][]visitmodel.Visit, len(studentIDOrder))
	for _, sid := range studentIDOrder {
		visitsByStudentID[sid] = []visitmodel.Visit{}
	}
	for _, lv := range lessonVisits {
		visitsByStudentID[lv.StudentID] = append(visitsByStudentID[lv.StudentID], lv)
	}

	var (
		studentsList        []studentmodel.Student
		balancesByStudentID map[uuid.UUID]txmodel.StudentBalance
	)

	if len(studentIDOrder) > 0 {
		g2, g2Ctx := errgroup.WithContext(ctx)
		if s.studentProvider != nil {
			g2.Go(func() error {
				stRows, err := s.studentProvider.GetStudentsByIDs(g2Ctx, studentIDOrder, ownerID)
				if err != nil {
					return err
				}
				stMap := make(map[uuid.UUID]studentmodel.Student, len(stRows))
				for _, sr := range stRows {
					stMap[sr.ID] = sr
				}
				studentsList = make([]studentmodel.Student, 0, len(studentIDOrder))
				for _, sid := range studentIDOrder {
					if st, ok := stMap[sid]; ok {
						studentsList = append(studentsList, st)
					}
				}
				return nil
			})
		}
		if s.transactionProvider != nil {
			g2.Go(func() error {
				bm, err := s.transactionProvider.GetStudentBalancesByStudentIDs(g2Ctx, studentIDOrder, ownerID)
				if err != nil {
					return err
				}
				balancesByStudentID = bm
				return nil
			})
		}
		if err := g2.Wait(); err != nil {
			logger.ErrorContext(ctx, "Failed to fetch students or balances for group report", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
			return domain.GroupReport{}, err
		}
	}

	logger.DebugContext(ctx, "Compiling group report data",
		"group_id", groupID.String(),
		"year", year,
		"month", month,
		"lessons_count", len(lessons),
		"students_count", len(studentsList),
		"user_id", ownerID.String(),
	)

	report := BuildGroupReport(grp, year, month, studentsList, lessons, visitsByStudentID, balancesByStudentID, time.Now().UTC())
	logger.InfoContext(ctx, "Group report generated successfully",
		"group_id", groupID.String(),
		"year", year,
		"month", month,
		"students_count", len(report.Students),
		"total_lessons", report.TotalLessons,
		"total_present", report.TotalPresent,
		"user_id", ownerID.String(),
	)
	return report, nil
}
