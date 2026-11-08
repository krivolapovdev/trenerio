package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/mapper"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/model"
	"github.com/stretchr/testify/assert"
)

func TestToVisitResponse(t *testing.T) {
	visitID := uuid.New()
	lessonID := uuid.New()
	studentID := uuid.New()
	now := time.Now()

	dbVisit := db.VisitsModuleVisit{
		ID:        visitID,
		Status:    db.VisitsModuleVisitStatusPRESENT,
		Type:      db.VisitsModuleVisitTypeREGULAR,
		LessonID:  lessonID,
		StudentID: studentID,
		CreatedAt: now,
	}

	resp := mapper.ToVisitResponse(dbVisit)
	assert.Equal(t, visitID, resp.ID)
	assert.Equal(t, model.VisitStatusPresent, resp.Status)
	assert.Equal(t, model.VisitTypeRegular, resp.Type)
	assert.Equal(t, lessonID, resp.LessonID)
	assert.Equal(t, studentID, resp.StudentID)
	assert.Equal(t, now, resp.CreatedAt)
}

func TestToVisitResponse_Variants(t *testing.T) {
	testCases := []struct {
		name       string
		dbStatus   db.VisitsModuleVisitStatus
		dbType     db.VisitsModuleVisitType
		wantStatus model.VisitStatus
		wantType   model.VisitType
	}{
		{
			name:       "Absent",
			dbStatus:   db.VisitsModuleVisitStatusABSENT,
			dbType:     db.VisitsModuleVisitTypeREGULAR,
			wantStatus: model.VisitStatusAbsent,
			wantType:   model.VisitTypeRegular,
		},
		{
			name:       "Excused",
			dbStatus:   db.VisitsModuleVisitStatusEXCUSED,
			dbType:     db.VisitsModuleVisitTypeREGULAR,
			wantStatus: model.VisitStatusExcused,
			wantType:   model.VisitTypeRegular,
		},
		{
			name:       "Free",
			dbStatus:   db.VisitsModuleVisitStatusPRESENT,
			dbType:     db.VisitsModuleVisitTypeFREE,
			wantStatus: model.VisitStatusPresent,
			wantType:   model.VisitTypeFree,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := db.VisitsModuleVisit{
				ID:     uuid.New(),
				Status: tc.dbStatus,
				Type:   tc.dbType,
			}
			resp := mapper.ToVisitResponse(v)
			assert.Equal(t, tc.wantStatus, resp.Status)
			assert.Equal(t, tc.wantType, resp.Type)
		})
	}
}

func TestVisitService_NewConstructors(t *testing.T) {
	s1 := NewService(nil, nil)
	assert.NotNil(t, s1)

	s2 := NewVisitService(nil, nil)
	assert.NotNil(t, s2)
}

func TestVisitService_EmptyIDQueries(t *testing.T) {
	svc := NewService(nil, nil)
	ownerID := uuid.New()

	t.Run("GetVisitsByLessonIDs_Empty", func(t *testing.T) {
		res, err := svc.GetVisitsByLessonIDs(t.Context(), []uuid.UUID{}, ownerID)
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("GetVisitsByStudentIDs_Empty", func(t *testing.T) {
		res, err := svc.GetVisitsByStudentIDs(t.Context(), []uuid.UUID{}, ownerID)
		assert.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestVisitStatus_IsPresent(t *testing.T) {
	assert.True(t, model.VisitStatusPresent.IsPresent())
	assert.False(t, model.VisitStatusAbsent.IsPresent())
	assert.False(t, model.VisitStatusExcused.IsPresent())
}

func TestFindVisitByLessonID(t *testing.T) {
	lesson1 := uuid.New()
	lesson2 := uuid.New()
	visits := []model.VisitResponse{
		{ID: uuid.New(), LessonID: lesson1, Status: model.VisitStatusPresent},
		{ID: uuid.New(), LessonID: lesson2, Status: model.VisitStatusAbsent},
	}

	found := model.FindVisitByLessonID(visits, lesson1)
	assert.NotNil(t, found)
	assert.Equal(t, lesson1, found.LessonID)
	assert.True(t, found.Status.IsPresent())

	notFound := model.FindVisitByLessonID(visits, uuid.New())
	assert.Nil(t, notFound)
}

func TestIsBillableVisit(t *testing.T) {
	assert.True(t, model.IsBillableVisit(model.VisitStatusPresent, model.VisitTypeRegular, false))
	assert.True(t, model.IsBillableVisit(model.VisitStatusAbsent, model.VisitTypeRegular, false))

	assert.False(t, model.IsBillableVisit(model.VisitStatusExcused, model.VisitTypeRegular, false))
	assert.False(t, model.IsBillableVisit(model.VisitStatusPresent, model.VisitTypeFree, false))
	assert.False(t, model.IsBillableVisit(model.VisitStatusAbsent, model.VisitTypeFree, false))
	assert.False(t, model.IsBillableVisit(model.VisitStatusPresent, model.VisitTypeRegular, true))
	assert.False(t, model.IsBillableVisit(model.VisitStatusAbsent, model.VisitTypeRegular, true))
}
