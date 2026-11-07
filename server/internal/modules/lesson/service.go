package lesson

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction"
	"github.com/krivolapovdev/trenerio/internal/modules/visit"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/shared/errors"
	"github.com/krivolapovdev/trenerio/internal/shared/logger"
	"github.com/krivolapovdev/trenerio/internal/shared/request"
	"github.com/krivolapovdev/trenerio/internal/shared/response"
	"github.com/krivolapovdev/trenerio/internal/shared/utils"
	"github.com/krivolapovdev/trenerio/internal/shared/validator"
)

type VisitProvider interface {
	GetVisitsByLessonID(ctx context.Context, lessonID, ownerID uuid.UUID) ([]response.VisitResponse, error)
	GetVisitsByLessonIDTx(ctx context.Context, qtx *db.Queries, lessonID, ownerID uuid.UUID) ([]response.VisitResponse, error)
	CreateVisits(ctx context.Context, lessonID uuid.UUID, studentVisits []request.StudentVisit, ownerID uuid.UUID) error
	CreateVisitsTx(ctx context.Context, qtx *db.Queries, lessonID uuid.UUID, studentVisits []request.StudentVisit, ownerID uuid.UUID) error
	UpdateVisitsByLessonID(ctx context.Context, lessonID uuid.UUID, requests []request.StudentVisit, ownerID uuid.UUID) error
	UpdateVisitsByLessonIDTx(ctx context.Context, qtx *db.Queries, lessonID uuid.UUID, requests []request.StudentVisit, ownerID uuid.UUID) ([]response.VisitResponse, error)
	RemoveVisitsByLessonID(ctx context.Context, lessonID, ownerID uuid.UUID) error
}

type StudentProvider interface {
	GetStudentsByIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) ([]response.StudentResponse, error)
}

type BalanceProvider interface {
	SyncLessonAttendanceBalancesTx(ctx context.Context, qtx *db.Queries, ownerID uuid.UUID, changes []transaction.AttendanceBalanceChange) error
}

type Service struct {
	db              *db.Queries
	sqlDB           *sql.DB
	visitProvider   VisitProvider
	studentProvider StudentProvider
	balanceProvider BalanceProvider
}

func NewService(
	queries *db.Queries,
	sqlDB *sql.DB,
	visitProvider VisitProvider,
	studentProvider StudentProvider,
	balanceProvider BalanceProvider,
) *Service {
	return &Service{
		db:              queries,
		sqlDB:           sqlDB,
		visitProvider:   visitProvider,
		studentProvider: studentProvider,
		balanceProvider: balanceProvider,
	}
}

func (s *Service) withTx(ctx context.Context, fn func(qtx *db.Queries) error) error {
	return db.WithTx(ctx, s.sqlDB, s.db, fn)
}

func (s *Service) GetLessonByID(ctx context.Context, lessonID, ownerID uuid.UUID) (response.LessonResponse, error) {
	logger.InfoContext(ctx, "Getting lesson by id", "lesson_id", lessonID.String(), "user_id", ownerID.String())

	l, err := s.db.FindLessonByIDAndOwnerID(ctx, db.FindLessonByIDAndOwnerIDParams{
		ID:      lessonID,
		OwnerID: ownerID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WarnContext(ctx, "Lesson not found", "lesson_id", lessonID.String(), "user_id", ownerID.String())
			return response.LessonResponse{}, pkgerrors.NewNotFound("Lesson", lessonID, ownerID)
		}

		logger.ErrorContext(ctx, "Failed to get lesson by id", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
		return response.LessonResponse{}, err
	}

	logger.DebugContext(ctx, "Retrieved lesson by id", "lesson_id", lessonID.String(), "user_id", ownerID.String())

	return toLessonResponse(l), nil
}

func (s *Service) GetLessonDetailsByID(ctx context.Context, lessonID, ownerID uuid.UUID) (response.LessonDetailsResponse, error) {
	logger.InfoContext(ctx, "Getting lesson details", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	lessonResp, err := s.GetLessonByID(ctx, lessonID, ownerID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get lesson for details", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
		return response.LessonDetailsResponse{}, err
	}

	var visitResponses []response.VisitResponse
	if s.visitProvider != nil {
		var err error
		visitResponses, err = s.visitProvider.GetVisitsByLessonID(ctx, lessonID, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch visits for lesson details", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return response.LessonDetailsResponse{}, err
		}
	}

	studentIDs := make([]uuid.UUID, 0, len(visitResponses))
	for _, vr := range visitResponses {
		studentIDs = append(studentIDs, vr.StudentID)
	}
	studentIDs = utils.UniqueUUIDs(studentIDs)

	students := make([]response.StudentResponse, 0, len(studentIDs))
	if len(studentIDs) > 0 && s.studentProvider != nil {
		stRows, err := s.studentProvider.GetStudentsByIDs(ctx, studentIDs, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch students for lesson details", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return response.LessonDetailsResponse{}, err
		}
		stMap := make(map[uuid.UUID]response.StudentResponse, len(stRows))
		for _, sr := range stRows {
			stMap[sr.ID] = sr
		}
		for _, sid := range studentIDs {
			if st, ok := stMap[sid]; ok {
				students = append(students, st)
			}
		}
	}

	logger.InfoContext(ctx, "Lesson details retrieved", "lesson_id", lessonID.String(), "visits_count", len(visitResponses), "students_count", len(students), "user_id", ownerID.String())
	return response.LessonDetailsResponse{
		ID:            lessonResp.ID,
		Date:          lessonResp.Date,
		CreatedAt:     lessonResp.CreatedAt,
		GroupID:       lessonResp.GroupID,
		StudentVisits: visitResponses,
		Students:      students,
	}, nil
}

func (s *Service) CreateLesson(ctx context.Context, req request.CreateLessonRequest, ownerID uuid.UUID) (response.LessonResponse, error) {
	logger.InfoContext(ctx, "Creating lesson", "date", req.Date.Format(time.RFC3339), "user_id", ownerID.String())
	if err := validator.ValidateDateNotFuture(req.Date, "date"); err != nil {
		logger.WarnContext(ctx, "Lesson date cannot be in the future", "date", req.Date.Format(time.RFC3339), "user_id", ownerID.String())
		return response.LessonResponse{}, err
	}

	var groupID uuid.NullUUID
	if req.GroupID != nil {
		groupID = uuid.NullUUID{UUID: *req.GroupID, Valid: true}
	}

	uniqueVisits := deduplicateStudentVisits(req.Students)

	var createdLesson db.LessonsModuleLesson
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		var err error
		createdLesson, err = qtx.CreateLesson(ctx, db.CreateLessonParams{
			OwnerID: ownerID,
			GroupID: groupID,
			Date:    req.Date,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create lesson record", "error", err, "user_id", ownerID.String())
			return err
		}

		if len(uniqueVisits) == 0 {
			return nil
		}

		if s.visitProvider != nil {
			if err := s.visitProvider.CreateVisitsTx(ctx, qtx, createdLesson.ID, uniqueVisits, ownerID); err != nil {
				logger.ErrorContext(ctx, "Failed to create visits for lesson", "lesson_id", createdLesson.ID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		if s.balanceProvider != nil {
			studentIDs := make([]uuid.UUID, 0, len(uniqueVisits))
			for _, sv := range uniqueVisits {
				studentIDs = append(studentIDs, sv.StudentID)
			}
			freeMap, err := s.fetchStudentFreeFlagsTx(ctx, qtx, studentIDs, ownerID)
			if err != nil {
				return err
			}

			changes := make([]transaction.AttendanceBalanceChange, 0, len(uniqueVisits))
			for _, sv := range uniqueVisits {
				changes = append(changes, transaction.AttendanceBalanceChange{
					StudentID:     sv.StudentID,
					OldBillable:   false,
					NewBillable:   visit.IsBillableVisit(sv.Status, sv.Type, freeMap[sv.StudentID]),
					NewLessonDate: createdLesson.Date,
				})
			}

			if err := s.balanceProvider.SyncLessonAttendanceBalancesTx(ctx, qtx, ownerID, changes); err != nil {
				logger.ErrorContext(ctx, "Failed to sync student balances on lesson creation", "lesson_id", createdLesson.ID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		return nil
	})
	if err != nil {
		return response.LessonResponse{}, err
	}

	logger.InfoContext(ctx, "Lesson created successfully", "lesson_id", createdLesson.ID.String(), "visits_count", len(uniqueVisits), "user_id", ownerID.String())
	return toLessonResponse(createdLesson), nil
}

func (s *Service) UpdateLesson(ctx context.Context, lessonID, ownerID uuid.UUID, req request.UpdateLessonRequest) (response.LessonResponse, error) {
	logger.InfoContext(ctx, "Updating lesson", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	if req.Date != nil {
		if err := validator.ValidateDateNotFuture(*req.Date, "date"); err != nil {
			logger.WarnContext(ctx, "Updated lesson date cannot be in the future", "date", req.Date.Format(time.RFC3339), "user_id", ownerID.String())
			return response.LessonResponse{}, err
		}
	}

	var updatedLesson db.LessonsModuleLesson
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		existingLesson, err := qtx.FindLessonByIDAndOwnerID(ctx, db.FindLessonByIDAndOwnerIDParams{
			ID:      lessonID,
			OwnerID: ownerID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Lesson to update not found", "lesson_id", lessonID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Lesson", lessonID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find lesson for update", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		oldDate := existingLesson.Date
		newDate := existingLesson.Date
		if req.Date != nil {
			newDate = *req.Date
		}

		updatedLesson, err = qtx.UpdateLesson(ctx, db.UpdateLessonParams{
			ID:      lessonID,
			OwnerID: ownerID,
			Date:    newDate,
		})
		if err != nil {
			logger.ErrorContext(ctx, "Failed to update lesson record", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		if req.Students != nil && s.visitProvider != nil {
			uniqueRequests := deduplicateStudentVisits(req.Students)
			previousVisits, err := s.visitProvider.UpdateVisitsByLessonIDTx(ctx, qtx, lessonID, uniqueRequests, ownerID)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to update visits for lesson", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
				return err
			}

			if s.balanceProvider != nil {
				oldMap := make(map[uuid.UUID]response.VisitResponse, len(previousVisits))
				studentIDSet := make([]uuid.UUID, 0, len(previousVisits)+len(uniqueRequests))
				for _, pv := range previousVisits {
					oldMap[pv.StudentID] = pv
					studentIDSet = append(studentIDSet, pv.StudentID)
				}

				newMap := make(map[uuid.UUID]request.StudentVisit, len(uniqueRequests))
				for _, nv := range uniqueRequests {
					newMap[nv.StudentID] = nv
					studentIDSet = append(studentIDSet, nv.StudentID)
				}
				allStudentIDs := utils.UniqueUUIDs(studentIDSet)

				freeMap, err := s.fetchStudentFreeFlagsTx(ctx, qtx, allStudentIDs, ownerID)
				if err != nil {
					return err
				}

				changes := make([]transaction.AttendanceBalanceChange, 0, len(allStudentIDs))
				for _, sid := range allStudentIDs {
					var oldBillable bool
					if ov, ok := oldMap[sid]; ok {
						oldBillable = visit.IsBillableVisit(ov.Status, ov.Type, freeMap[sid])
					}
					var newBillable bool
					if nv, ok := newMap[sid]; ok {
						newBillable = visit.IsBillableVisit(nv.Status, nv.Type, freeMap[sid])
					}
					changes = append(changes, transaction.AttendanceBalanceChange{
						StudentID:     sid,
						OldBillable:   oldBillable,
						OldLessonDate: oldDate,
						NewBillable:   newBillable,
						NewLessonDate: newDate,
					})
				}

				if err := s.balanceProvider.SyncLessonAttendanceBalancesTx(ctx, qtx, ownerID, changes); err != nil {
					logger.ErrorContext(ctx, "Failed to sync student balances on lesson update", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
					return err
				}
			}
		} else if req.Students == nil && !newDate.Equal(oldDate) && s.visitProvider != nil && s.balanceProvider != nil {
			existingVisits, err := s.visitProvider.GetVisitsByLessonIDTx(ctx, qtx, lessonID, ownerID)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to fetch visits for lesson date change balance sync", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
			if len(existingVisits) > 0 {
				studentIDs := make([]uuid.UUID, 0, len(existingVisits))
				for _, ev := range existingVisits {
					studentIDs = append(studentIDs, ev.StudentID)
				}
				studentIDs = utils.UniqueUUIDs(studentIDs)

				freeMap, err := s.fetchStudentFreeFlagsTx(ctx, qtx, studentIDs, ownerID)
				if err != nil {
					return err
				}

				changes := make([]transaction.AttendanceBalanceChange, 0, len(existingVisits))
				for _, ev := range existingVisits {
					billable := visit.IsBillableVisit(ev.Status, ev.Type, freeMap[ev.StudentID])
					changes = append(changes, transaction.AttendanceBalanceChange{
						StudentID:     ev.StudentID,
						OldBillable:   billable,
						OldLessonDate: oldDate,
						NewBillable:   billable,
						NewLessonDate: newDate,
					})
				}

				if err := s.balanceProvider.SyncLessonAttendanceBalancesTx(ctx, qtx, ownerID, changes); err != nil {
					logger.ErrorContext(ctx, "Failed to sync student balances on lesson date update", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return response.LessonResponse{}, err
	}

	logger.InfoContext(ctx, "Lesson updated successfully", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	return toLessonResponse(updatedLesson), nil
}

func (s *Service) DeleteLesson(ctx context.Context, lessonID, ownerID uuid.UUID) error {
	logger.InfoContext(ctx, "Deleting lesson", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	err := s.withTx(ctx, func(qtx *db.Queries) error {
		existingLesson, err := qtx.FindLessonByIDAndOwnerID(ctx, db.FindLessonByIDAndOwnerIDParams{
			ID:      lessonID,
			OwnerID: ownerID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				logger.WarnContext(ctx, "Lesson to delete not found", "lesson_id", lessonID.String(), "user_id", ownerID.String())
				return pkgerrors.NewNotFound("Lesson", lessonID, ownerID)
			}
			logger.ErrorContext(ctx, "Failed to find lesson for deletion", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		var existingVisits []response.VisitResponse
		if s.balanceProvider != nil {
			if s.visitProvider != nil {
				existingVisits, err = s.visitProvider.GetVisitsByLessonIDTx(ctx, qtx, lessonID, ownerID)
				if err != nil {
					logger.ErrorContext(ctx, "Failed to fetch visits before lesson deletion", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
					return err
				}
			} else if qtx != nil {
				rows, err := qtx.FindVisitsByLessonIDAndOwnerID(ctx, db.FindVisitsByLessonIDAndOwnerIDParams{
					LessonID: lessonID,
					OwnerID:  ownerID,
				})
				if err != nil {
					return err
				}
				for _, r := range rows {
					existingVisits = append(existingVisits, response.VisitResponse{
						ID:        r.ID,
						LessonID:  r.LessonID,
						StudentID: r.StudentID,
						Status:    visit.VisitStatus(r.Status),
						Type:      visit.VisitType(r.Type),
						CreatedAt: r.CreatedAt,
					})
				}
			}
		}

		if err := qtx.DeleteLessonByIDAndOwnerID(ctx, db.DeleteLessonByIDAndOwnerIDParams{
			ID:      lessonID,
			OwnerID: ownerID,
		}); err != nil {
			logger.ErrorContext(ctx, "Failed to delete lesson record", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		if len(existingVisits) > 0 && s.balanceProvider != nil {
			studentIDs := make([]uuid.UUID, 0, len(existingVisits))
			for _, ev := range existingVisits {
				studentIDs = append(studentIDs, ev.StudentID)
			}
			studentIDs = utils.UniqueUUIDs(studentIDs)

			freeMap, err := s.fetchStudentFreeFlagsTx(ctx, qtx, studentIDs, ownerID)
			if err != nil {
				return err
			}

			changes := make([]transaction.AttendanceBalanceChange, 0, len(existingVisits))
			for _, ev := range existingVisits {
				changes = append(changes, transaction.AttendanceBalanceChange{
					StudentID:     ev.StudentID,
					OldBillable:   visit.IsBillableVisit(ev.Status, ev.Type, freeMap[ev.StudentID]),
					OldLessonDate: existingLesson.Date,
					NewBillable:   false,
				})
			}

			if err := s.balanceProvider.SyncLessonAttendanceBalancesTx(ctx, qtx, ownerID, changes); err != nil {
				logger.ErrorContext(ctx, "Failed to sync student balances on lesson deletion", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Lesson deleted successfully", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	return nil
}

func (s *Service) GetLessonsByGroupIDAndDateRange(ctx context.Context, groupID, ownerID uuid.UUID, from, to time.Time) ([]response.LessonResponse, error) {
	logger.InfoContext(ctx, "Getting lessons by group id and date range", "group_id", groupID.String(), "user_id", ownerID.String())
	rows, err := s.db.FindLessonsByGroupIDAndDateBetween(ctx, db.FindLessonsByGroupIDAndDateBetweenParams{
		GroupID: uuid.NullUUID{UUID: groupID, Valid: true},
		OwnerID: ownerID,
		Date:    from,
		Date_2:  to,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get lessons by group id and date range", "group_id", groupID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}
	lessons := make([]response.LessonResponse, 0, len(rows))
	for _, r := range rows {
		lessons = append(lessons, toLessonResponse(r))
	}
	logger.DebugContext(ctx, "Retrieved lessons by group id and date range", "group_id", groupID.String(), "count", len(lessons), "user_id", ownerID.String())
	return lessons, nil
}

func (s *Service) GetLessonsByIDs(ctx context.Context, lessonIDs []uuid.UUID, ownerID uuid.UUID) ([]response.LessonResponse, error) {
	if len(lessonIDs) == 0 {
		logger.DebugContext(ctx, "No lesson IDs provided to get lessons", "user_id", ownerID.String())
		return []response.LessonResponse{}, nil
	}
	rows, err := s.db.FindLessonsByIDsAndOwnerID(ctx, db.FindLessonsByIDsAndOwnerIDParams{
		OwnerID: ownerID,
		Column2: lessonIDs,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get lessons by IDs", "count", len(lessonIDs), "error", err, "user_id", ownerID.String())
		return nil, err
	}
	lessons := make([]response.LessonResponse, 0, len(rows))
	for _, r := range rows {
		lessons = append(lessons, toLessonResponse(r))
	}
	logger.DebugContext(ctx, "Retrieved lessons by IDs", "count", len(lessons), "user_id", ownerID.String())
	return lessons, nil
}

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

func deduplicateStudentVisits(visits []request.StudentVisit) []request.StudentVisit {
	if len(visits) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]int, len(visits))
	unique := make([]request.StudentVisit, 0, len(visits))
	for _, sv := range visits {
		if idx, exists := seen[sv.StudentID]; exists {
			unique[idx] = sv
		} else {
			seen[sv.StudentID] = len(unique)
			unique = append(unique, sv)
		}
	}
	return unique
}
