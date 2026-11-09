package http

import (
	"bytes"
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/harness"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLessonService struct {
	getLessonByIDFn        func(ctx context.Context, lessonID, ownerID uuid.UUID) (domain.Lesson, error)
	getLessonDetailsByIDFn func(ctx context.Context, lessonID, ownerID uuid.UUID) (domain.LessonDetails, error)
	createLessonFn         func(ctx context.Context, groupID *uuid.UUID, date time.Time, studentVisits []visitmodel.StudentVisit, ownerID uuid.UUID) (domain.Lesson, error)
	updateLessonFn         func(ctx context.Context, lessonID, ownerID uuid.UUID, date *time.Time, studentVisits []visitmodel.StudentVisit) (domain.Lesson, error)
	deleteLessonFn         func(ctx context.Context, lessonID, ownerID uuid.UUID) error
}

func (m *mockLessonService) GetLessonByID(ctx context.Context, lessonID, ownerID uuid.UUID) (domain.Lesson, error) {
	if m.getLessonByIDFn != nil {
		return m.getLessonByIDFn(ctx, lessonID, ownerID)
	}
	return domain.Lesson{}, nil
}

func (m *mockLessonService) GetLessonDetailsByID(ctx context.Context, lessonID, ownerID uuid.UUID) (domain.LessonDetails, error) {
	if m.getLessonDetailsByIDFn != nil {
		return m.getLessonDetailsByIDFn(ctx, lessonID, ownerID)
	}
	return domain.LessonDetails{}, nil
}

func (m *mockLessonService) CreateLesson(ctx context.Context, groupID *uuid.UUID, date time.Time, studentVisits []visitmodel.StudentVisit, ownerID uuid.UUID) (domain.Lesson, error) {
	if m.createLessonFn != nil {
		return m.createLessonFn(ctx, groupID, date, studentVisits, ownerID)
	}
	return domain.Lesson{}, nil
}

func (m *mockLessonService) UpdateLesson(ctx context.Context, lessonID, ownerID uuid.UUID, date *time.Time, studentVisits []visitmodel.StudentVisit) (domain.Lesson, error) {
	if m.updateLessonFn != nil {
		return m.updateLessonFn(ctx, lessonID, ownerID, date, studentVisits)
	}
	return domain.Lesson{}, nil
}

func (m *mockLessonService) DeleteLesson(ctx context.Context, lessonID, ownerID uuid.UUID) error {
	if m.deleteLessonFn != nil {
		return m.deleteLessonFn(ctx, lessonID, ownerID)
	}
	return nil
}

func setupLessonRouter(svc lessonService, user authmodel.JwtUser) *gin.Engine {
	return harness.SetupTestRouter(user, func(rg *gin.RouterGroup) {
		NewHandler(svc).RegisterRoutes(rg)
	})
}

func TestLessonHandler_GetLessonDetails(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	lessonID := uuid.New()

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/lessons/"+lessonID.String()+"/details", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidLessonID", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/lessons/invalid-uuid/details", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSvc := &mockLessonService{
			getLessonDetailsByIDFn: func(ctx context.Context, lID, ownerID uuid.UUID) (domain.LessonDetails, error) {
				return domain.LessonDetails{}, pkgerrors.NewNotFound("Lesson", lID, ownerID)
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/lessons/"+lessonID.String()+"/details", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		mockSvc := &mockLessonService{
			getLessonDetailsByIDFn: func(ctx context.Context, lID, ownerID uuid.UUID) (domain.LessonDetails, error) {
				assert.Equal(t, lessonID, lID)
				assert.Equal(t, testUser.ID, ownerID)
				return domain.LessonDetails{
					ID:   lessonID,
					Date: time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC),
				}, nil
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/lessons/"+lessonID.String()+"/details", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
		var res LessonDetailsResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, lessonID, res.ID)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockLessonService{
			getLessonDetailsByIDFn: func(ctx context.Context, lID, ownerID uuid.UUID) (domain.LessonDetails, error) {
				return domain.LessonDetails{}, assert.AnError
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/lessons/"+lessonID.String()+"/details", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestLessonHandler_CreateLesson(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/lessons", bytes.NewBufferString(`{"date":"2026-10-17T10:00:00Z"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/lessons", bytes.NewBufferString(`{invalid`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("FutureDate", func(t *testing.T) {
		futureDate := time.Now().UTC().Add(48 * time.Hour)
		r := setupLessonRouter(&mockLessonService{}, testUser)
		w := httptest.NewRecorder()
		payload := `{"date":"` + futureDate.Format(time.RFC3339) + `"}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/lessons", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("MissingRequiredDate", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/lessons", bytes.NewBufferString(`{"students":[]}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		lessonDate := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
		expectedID := uuid.New()

		mockSvc := &mockLessonService{
			createLessonFn: func(ctx context.Context, groupID *uuid.UUID, date time.Time, studentVisits []visitmodel.StudentVisit, ownerID uuid.UUID) (domain.Lesson, error) {
				assert.Equal(t, testUser.ID, ownerID)
				assert.True(t, lessonDate.Equal(date))
				return domain.Lesson{ID: expectedID, Date: date}, nil
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		payload := `{"date":"` + lessonDate.Format(time.RFC3339) + `"}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/lessons", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusCreated, w.Code)
	})

	t.Run("NormalizesExcusedVisitsToFree", func(t *testing.T) {
		lessonDate := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
		expectedID := uuid.New()
		studentID := uuid.New()

		mockSvc := &mockLessonService{
			createLessonFn: func(ctx context.Context, groupID *uuid.UUID, date time.Time, studentVisits []visitmodel.StudentVisit, ownerID uuid.UUID) (domain.Lesson, error) {
				assert.Len(t, studentVisits, 1)
				assert.Equal(t, visitmodel.VisitStatusExcused, studentVisits[0].Status)
				assert.Equal(t, visitmodel.VisitTypeFree, studentVisits[0].Type)
				return domain.Lesson{ID: expectedID, Date: date}, nil
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		payload := `{"date":"` + lessonDate.Format(time.RFC3339) + `","students":[{"studentId":"` + studentID.String() + `","status":"EXCUSED","type":"REGULAR"}]}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/lessons", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusCreated, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		lessonDate := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
		mockSvc := &mockLessonService{
			createLessonFn: func(ctx context.Context, groupID *uuid.UUID, date time.Time, studentVisits []visitmodel.StudentVisit, ownerID uuid.UUID) (domain.Lesson, error) {
				return domain.Lesson{}, assert.AnError
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		payload := `{"date":"` + lessonDate.Format(time.RFC3339) + `"}`
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/lessons", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestLessonHandler_UpdateLesson(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	lessonID := uuid.New()

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/lessons/"+lessonID.String(), bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidLessonID", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/lessons/bad-id", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/lessons/"+lessonID.String(), bytes.NewBufferString(`{invalid`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("FutureDate", func(t *testing.T) {
		futureDate := time.Now().UTC().Add(48 * time.Hour)
		r := setupLessonRouter(&mockLessonService{}, testUser)
		w := httptest.NewRecorder()
		payload := `{"date":"` + futureDate.Format(time.RFC3339) + `"}`
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/lessons/"+lessonID.String(), bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Success_WithoutDate", func(t *testing.T) {
		mockSvc := &mockLessonService{
			updateLessonFn: func(ctx context.Context, lID, ownerID uuid.UUID, date *time.Time, studentVisits []visitmodel.StudentVisit) (domain.Lesson, error) {
				assert.Equal(t, lessonID, lID)
				assert.Equal(t, testUser.ID, ownerID)
				assert.Nil(t, date)
				return domain.Lesson{ID: lessonID}, nil
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/lessons/"+lessonID.String(), bytes.NewBufferString(`{"students":[]}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		newDate := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
		mockSvc := &mockLessonService{
			updateLessonFn: func(ctx context.Context, lID, ownerID uuid.UUID, date *time.Time, studentVisits []visitmodel.StudentVisit) (domain.Lesson, error) {
				assert.Equal(t, lessonID, lID)
				assert.Equal(t, testUser.ID, ownerID)
				require.NotNil(t, date)
				assert.True(t, newDate.Equal(*date))
				return domain.Lesson{ID: lessonID, Date: *date}, nil
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		payload := `{"date":"` + newDate.Format(time.RFC3339) + `"}`
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/lessons/"+lessonID.String(), bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
	})

	t.Run("NormalizesExcusedVisitsToFree", func(t *testing.T) {
		lessonID := uuid.New()
		studentID := uuid.New()

		mockSvc := &mockLessonService{
			updateLessonFn: func(ctx context.Context, lID, ownerID uuid.UUID, date *time.Time, studentVisits []visitmodel.StudentVisit) (domain.Lesson, error) {
				assert.Len(t, studentVisits, 1)
				assert.Equal(t, visitmodel.VisitStatusExcused, studentVisits[0].Status)
				assert.Equal(t, visitmodel.VisitTypeFree, studentVisits[0].Type)
				return domain.Lesson{ID: lessonID}, nil
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		payload := `{"students":[{"studentId":"` + studentID.String() + `","status":"EXCUSED","type":"REGULAR"}]}`
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/lessons/"+lessonID.String(), bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSvc := &mockLessonService{
			updateLessonFn: func(ctx context.Context, lID, ownerID uuid.UUID, date *time.Time, studentVisits []visitmodel.StudentVisit) (domain.Lesson, error) {
				return domain.Lesson{}, pkgerrors.NewNotFound("Lesson", lID, ownerID)
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/lessons/"+lessonID.String(), bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockLessonService{
			updateLessonFn: func(ctx context.Context, lID, ownerID uuid.UUID, date *time.Time, studentVisits []visitmodel.StudentVisit) (domain.Lesson, error) {
				return domain.Lesson{}, assert.AnError
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/lessons/"+lessonID.String(), bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestLessonHandler_DeleteLesson(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	lessonID := uuid.New()

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/lessons/"+lessonID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidLessonID", func(t *testing.T) {
		r := setupLessonRouter(&mockLessonService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/lessons/bad-id", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		called := false
		mockSvc := &mockLessonService{
			deleteLessonFn: func(ctx context.Context, lID, ownerID uuid.UUID) error {
				assert.Equal(t, lessonID, lID)
				assert.Equal(t, testUser.ID, ownerID)
				called = true
				return nil
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/lessons/"+lessonID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNoContent, w.Code)
		assert.True(t, called)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSvc := &mockLessonService{
			deleteLessonFn: func(ctx context.Context, lID, ownerID uuid.UUID) error {
				return pkgerrors.NewNotFound("Lesson", lID, ownerID)
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/lessons/"+lessonID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockLessonService{
			deleteLessonFn: func(ctx context.Context, lID, ownerID uuid.UUID) error {
				return assert.AnError
			},
		}

		r := setupLessonRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/lessons/"+lessonID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}
