package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/utils"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

func (s *Service) GetLessonDetailsByID(ctx context.Context, lessonID, ownerID uuid.UUID) (domain.LessonDetails, error) {
	logger.InfoContext(ctx, "Getting lesson details", "lesson_id", lessonID.String(), "user_id", ownerID.String())
	lessonResp, err := s.GetLessonByID(ctx, lessonID, ownerID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to get lesson for details", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
		return domain.LessonDetails{}, err
	}

	var visitResponses []visitmodel.Visit
	if s.visitProvider != nil {
		var err error
		visitResponses, err = s.visitProvider.GetVisitsByLessonID(ctx, lessonID, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch visits for lesson details", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return domain.LessonDetails{}, err
		}
	}

	studentIDs := make([]uuid.UUID, 0, len(visitResponses))
	for _, vr := range visitResponses {
		studentIDs = append(studentIDs, vr.StudentID)
	}
	studentIDs = utils.UniqueUUIDs(studentIDs)

	students := make([]studentmodel.Student, 0, len(studentIDs))
	if len(studentIDs) > 0 && s.studentProvider != nil {
		stRows, err := s.studentProvider.GetStudentsByIDs(ctx, studentIDs, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to fetch students for lesson details", "lesson_id", lessonID.String(), "error", err, "user_id", ownerID.String())
			return domain.LessonDetails{}, err
		}
		stMap := make(map[uuid.UUID]studentmodel.Student, len(stRows))
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
	return domain.LessonDetails{
		ID:            lessonResp.ID,
		Date:          lessonResp.Date,
		CreatedAt:     lessonResp.CreatedAt,
		GroupID:       lessonResp.GroupID,
		StudentVisits: visitResponses,
		Students:      students,
	}, nil
}
