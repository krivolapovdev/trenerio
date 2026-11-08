package service

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

func (s *Service) GetStudentVisits(ctx context.Context, studentID, ownerID uuid.UUID, from, to *time.Time) ([]StudentVisitWithLesson, error) {
	logger.InfoContext(ctx, "Getting student visits", "student_id", studentID.String(), "user_id", ownerID.String())
	if _, err := s.GetStudentDetailsByID(ctx, studentID, ownerID); err != nil {
		logger.WarnContext(ctx, "Student not found for visits", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return nil, err
	}

	var vRows []visitmodel.Visit
	if s.visitProvider != nil {
		var err error
		vRows, err = s.visitProvider.GetVisitsByStudentID(ctx, studentID, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch student visits", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return nil, err
		}
	}
	if len(vRows) == 0 {
		return []StudentVisitWithLesson{}, nil
	}

	lessonIDs := make([]uuid.UUID, 0, len(vRows))
	for _, vr := range vRows {
		lessonIDs = append(lessonIDs, vr.LessonID)
	}

	lessonsMap := make(map[uuid.UUID]lessonmodel.Lesson, len(lessonIDs))
	if len(lessonIDs) > 0 && s.lessonProvider != nil {
		lRows, err := s.lessonProvider.GetLessonsByIDs(ctx, lessonIDs, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch lessons for student visits", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return nil, err
		}
		for _, lr := range lRows {
			lessonsMap[lr.ID] = lr
		}
	}

	visitsWithLessons := make([]StudentVisitWithLesson, 0, len(vRows))
	for _, vr := range vRows {
		lResp, ok := lessonsMap[vr.LessonID]
		if !ok {
			continue
		}
		if from != nil && lResp.Date.Before(*from) {
			continue
		}
		if to != nil && lResp.Date.After(*to) {
			continue
		}
		visitsWithLessons = append(visitsWithLessons, StudentVisitWithLesson{
			Visit:  vr,
			Lesson: lResp,
		})
	}

	sort.Slice(visitsWithLessons, func(i, j int) bool {
		return visitsWithLessons[i].Lesson.Date.After(visitsWithLessons[j].Lesson.Date)
	})

	logger.InfoContext(ctx, "Retrieved student visits",
		"student_id", studentID.String(),
		"total_visits", len(vRows),
		"filtered_visits", len(visitsWithLessons),
		"user_id", ownerID.String(),
	)
	return visitsWithLessons, nil
}
