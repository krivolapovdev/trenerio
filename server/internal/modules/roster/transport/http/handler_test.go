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
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRosterService struct {
	getAllRostersFn         func(ctx context.Context, ownerID uuid.UUID) ([]domain.Roster, error)
	getRosterByIDFn         func(ctx context.Context, rosterID, ownerID uuid.UUID) (domain.Roster, error)
	getRosterStudentsFn     func(ctx context.Context, rosterID, ownerID uuid.UUID) ([]studentmodel.Student, error)
	createRosterFn          func(ctx context.Context, rosterToCreate domain.Roster, studentIDs []uuid.UUID, ownerID uuid.UUID) (domain.Roster, error)
	updateRosterFn          func(ctx context.Context, rosterID, ownerID uuid.UUID, updates map[string]any) (domain.Roster, error)
	deleteRosterFn          func(ctx context.Context, rosterID, ownerID uuid.UUID) error
	addStudentsFn           func(ctx context.Context, rosterID, ownerID uuid.UUID, studentIDs []uuid.UUID, note *string) (domain.Roster, error)
	updateParticipantNoteFn func(ctx context.Context, rosterID, studentID, ownerID uuid.UUID, note *string) error
	removeStudentFn         func(ctx context.Context, rosterID, studentID, ownerID uuid.UUID) error
}

func (m *mockRosterService) GetAllRosters(ctx context.Context, ownerID uuid.UUID) ([]domain.Roster, error) {
	if m.getAllRostersFn != nil {
		return m.getAllRostersFn(ctx, ownerID)
	}
	return nil, nil
}

func (m *mockRosterService) GetRosterByID(ctx context.Context, rosterID, ownerID uuid.UUID) (domain.Roster, error) {
	if m.getRosterByIDFn != nil {
		return m.getRosterByIDFn(ctx, rosterID, ownerID)
	}
	return domain.Roster{}, nil
}

func (m *mockRosterService) GetRosterStudents(ctx context.Context, rosterID, ownerID uuid.UUID) ([]studentmodel.Student, error) {
	if m.getRosterStudentsFn != nil {
		return m.getRosterStudentsFn(ctx, rosterID, ownerID)
	}
	return nil, nil
}

func (m *mockRosterService) CreateRoster(ctx context.Context, rosterToCreate domain.Roster, studentIDs []uuid.UUID, ownerID uuid.UUID) (domain.Roster, error) {
	if m.createRosterFn != nil {
		return m.createRosterFn(ctx, rosterToCreate, studentIDs, ownerID)
	}
	return domain.Roster{}, nil
}

func (m *mockRosterService) UpdateRoster(ctx context.Context, rosterID, ownerID uuid.UUID, updates map[string]any) (domain.Roster, error) {
	if m.updateRosterFn != nil {
		return m.updateRosterFn(ctx, rosterID, ownerID, updates)
	}
	return domain.Roster{}, nil
}

func (m *mockRosterService) DeleteRoster(ctx context.Context, rosterID, ownerID uuid.UUID) error {
	if m.deleteRosterFn != nil {
		return m.deleteRosterFn(ctx, rosterID, ownerID)
	}
	return nil
}

func (m *mockRosterService) AddStudents(ctx context.Context, rosterID, ownerID uuid.UUID, studentIDs []uuid.UUID, note *string) (domain.Roster, error) {
	if m.addStudentsFn != nil {
		return m.addStudentsFn(ctx, rosterID, ownerID, studentIDs, note)
	}
	return domain.Roster{}, nil
}

func (m *mockRosterService) UpdateParticipantNote(ctx context.Context, rosterID, studentID, ownerID uuid.UUID, note *string) error {
	if m.updateParticipantNoteFn != nil {
		return m.updateParticipantNoteFn(ctx, rosterID, studentID, ownerID, note)
	}
	return nil
}

func (m *mockRosterService) RemoveStudent(ctx context.Context, rosterID, studentID, ownerID uuid.UUID) error {
	if m.removeStudentFn != nil {
		return m.removeStudentFn(ctx, rosterID, studentID, ownerID)
	}
	return nil
}

func setupRosterRouter(svc rosterService, user authmodel.JwtUser) *gin.Engine {
	return harness.SetupTestRouter(user, func(rg *gin.RouterGroup) {
		NewHandler(svc).RegisterRoutes(rg)
	})
}

func TestRosterHandler_GetAllRosters(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		expected := []domain.Roster{
			{ID: uuid.New(), Name: "Competition A", CountOfStudents: 3},
		}
		mockSvc := &mockRosterService{
			getAllRostersFn: func(ctx context.Context, ownerID uuid.UUID) ([]domain.Roster, error) {
				assert.Equal(t, testUser.ID, ownerID)
				return expected, nil
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
		var actual []domain.Roster
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		require.NoError(t, err)
		assert.Equal(t, expected[0].ID, actual[0].ID)
		assert.Equal(t, "Competition A", actual[0].Name)
		assert.Equal(t, int64(3), actual[0].CountOfStudents)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockSvc := &mockRosterService{
			getAllRostersFn: func(ctx context.Context, ownerID uuid.UUID) ([]domain.Roster, error) {
				return nil, assert.AnError
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestRosterHandler_GetRosterByID(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	rosterID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		expected := domain.Roster{
			ID:              rosterID,
			Name:            "Competition A",
			CountOfStudents: 1,
			Students: []studentmodel.Student{
				{ID: uuid.New(), FullName: "John Doe"},
			},
		}
		mockSvc := &mockRosterService{
			getRosterByIDFn: func(ctx context.Context, rID, ownerID uuid.UUID) (domain.Roster, error) {
				assert.Equal(t, rosterID, rID)
				assert.Equal(t, testUser.ID, ownerID)
				return expected, nil
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/"+rosterID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
		var actual domain.Roster
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		require.NoError(t, err)
		assert.Equal(t, rosterID, actual.ID)
		assert.Equal(t, "Competition A", actual.Name)
		assert.Len(t, actual.Students, 1)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/"+rosterID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidRosterID", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/not-a-uuid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSvc := &mockRosterService{
			getRosterByIDFn: func(ctx context.Context, rID, ownerID uuid.UUID) (domain.Roster, error) {
				return domain.Roster{}, pkgerrors.NewNotFound("Roster", rID, ownerID)
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/"+rosterID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockRosterService{
			getRosterByIDFn: func(ctx context.Context, rID, ownerID uuid.UUID) (domain.Roster, error) {
				return domain.Roster{}, assert.AnError
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/"+rosterID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestRosterHandler_GetRosterStudents(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	rosterID := uuid.New()

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/"+rosterID.String()+"/students", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidRosterID", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/invalid-id/students", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		expectedStudents := []studentmodel.Student{
			{ID: uuid.New(), FullName: "Alice Smith"},
			{ID: uuid.New(), FullName: "Bob Jones"},
		}
		mockSvc := &mockRosterService{
			getRosterStudentsFn: func(ctx context.Context, rID, ownerID uuid.UUID) ([]studentmodel.Student, error) {
				assert.Equal(t, rosterID, rID)
				assert.Equal(t, testUser.ID, ownerID)
				return expectedStudents, nil
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/"+rosterID.String()+"/students", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
		var actual []StudentResponse
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		require.NoError(t, err)
		assert.Len(t, actual, 2)
		assert.Equal(t, expectedStudents[0].FullName, actual[0].FullName)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSvc := &mockRosterService{
			getRosterStudentsFn: func(ctx context.Context, rID, ownerID uuid.UUID) ([]studentmodel.Student, error) {
				return nil, pkgerrors.NewNotFound("Roster", rID, ownerID)
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/"+rosterID.String()+"/students", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockRosterService{
			getRosterStudentsFn: func(ctx context.Context, rID, ownerID uuid.UUID) ([]studentmodel.Student, error) {
				return nil, assert.AnError
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/rosters/"+rosterID.String()+"/students", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestRosterHandler_CreateRoster(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}

	t.Run("Success", func(t *testing.T) {
		now := time.Now().UTC()
		note := "Championship"
		reqBody := CreateRosterRequest{
			Name: "Competition A",
			Date: &now,
			Note: &note,
		}

		mockSvc := &mockRosterService{
			createRosterFn: func(ctx context.Context, rosterToCreate domain.Roster, studentIDs []uuid.UUID, ownerID uuid.UUID) (domain.Roster, error) {
				assert.Equal(t, "Competition A", rosterToCreate.Name)
				assert.Equal(t, &note, rosterToCreate.Note)
				assert.Equal(t, testUser.ID, ownerID)
				return domain.Roster{
					ID:   uuid.New(),
					Name: rosterToCreate.Name,
					Date: rosterToCreate.Date,
					Note: rosterToCreate.Note,
				}, nil
			},
		}

		bodyBytes, _ := json.Marshal(reqBody)
		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/rosters", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusCreated, w.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/rosters", bytes.NewBufferString(`{"name":"A"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("Validation_BlankName", func(t *testing.T) {
		reqBody := map[string]any{
			"name": "   ",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/rosters", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockSvc := &mockRosterService{
			createRosterFn: func(ctx context.Context, rosterToCreate domain.Roster, studentIDs []uuid.UUID, ownerID uuid.UUID) (domain.Roster, error) {
				return domain.Roster{}, assert.AnError
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/rosters", bytes.NewBufferString(`{"name":"Valid Name"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestRosterHandler_UpdateRoster(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	rosterID := uuid.New()

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String(), bytes.NewBufferString(`{"name":"Updated"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidRosterID", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/bad-id", bytes.NewBufferString(`{"name":"Updated"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String(), bytes.NewBufferString(`{bad-json`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Validation_BlankName", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String(), bytes.NewBufferString(`{"name":"   "}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		mockSvc := &mockRosterService{
			updateRosterFn: func(ctx context.Context, rID, ownerID uuid.UUID, updates map[string]any) (domain.Roster, error) {
				assert.Equal(t, rosterID, rID)
				assert.Equal(t, "Competition B", updates["name"])
				return domain.Roster{
					ID:   rID,
					Name: "Competition B",
				}, nil
			},
		}

		updates := map[string]any{"name": "Competition B"}
		bodyBytes, _ := json.Marshal(updates)
		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSvc := &mockRosterService{
			updateRosterFn: func(ctx context.Context, rID, ownerID uuid.UUID, updates map[string]any) (domain.Roster, error) {
				return domain.Roster{}, pkgerrors.NewNotFound("Roster", rID, ownerID)
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String(), bytes.NewBufferString(`{"name":"Valid Name"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockRosterService{
			updateRosterFn: func(ctx context.Context, rID, ownerID uuid.UUID, updates map[string]any) (domain.Roster, error) {
				return domain.Roster{}, assert.AnError
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String(), bytes.NewBufferString(`{"name":"Valid Name"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestRosterHandler_DeleteRoster(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	rosterID := uuid.New()

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/"+rosterID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidRosterID", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/invalid-id", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		mockSvc := &mockRosterService{
			deleteRosterFn: func(ctx context.Context, rID, ownerID uuid.UUID) error {
				assert.Equal(t, rosterID, rID)
				return nil
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/"+rosterID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNoContent, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSvc := &mockRosterService{
			deleteRosterFn: func(ctx context.Context, rID, ownerID uuid.UUID) error {
				return pkgerrors.NewNotFound("Roster", rID, ownerID)
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/"+rosterID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockRosterService{
			deleteRosterFn: func(ctx context.Context, rID, ownerID uuid.UUID) error {
				return assert.AnError
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/"+rosterID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestRosterHandler_AddAndRemoveStudents(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "coach@example.com"}
	rosterID := uuid.New()
	studentID := uuid.New()

	t.Run("AddStudents_Unauthorized", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/rosters/"+rosterID.String()+"/students", bytes.NewBufferString(`{"studentIds":["`+studentID.String()+`"]}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("AddStudents_InvalidRosterID", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/rosters/bad-id/students", bytes.NewBufferString(`{"studentIds":["`+studentID.String()+`"]}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("AddStudents_InvalidJSON", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/rosters/"+rosterID.String()+"/students", bytes.NewBufferString(`{bad`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("AddStudents_ServiceError", func(t *testing.T) {
		mockSvc := &mockRosterService{
			addStudentsFn: func(ctx context.Context, rID, ownerID uuid.UUID, studentIDs []uuid.UUID, note *string) (domain.Roster, error) {
				return domain.Roster{}, assert.AnError
			},
		}

		reqBody := AddRosterStudentsRequest{
			StudentIds: []uuid.UUID{studentID},
		}
		bodyBytes, _ := json.Marshal(reqBody)
		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/rosters/"+rosterID.String()+"/students", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})

	t.Run("AddStudents_Success", func(t *testing.T) {
		mockSvc := &mockRosterService{
			addStudentsFn: func(ctx context.Context, rID, ownerID uuid.UUID, studentIDs []uuid.UUID, note *string) (domain.Roster, error) {
				assert.Equal(t, rosterID, rID)
				assert.Equal(t, []uuid.UUID{studentID}, studentIDs)
				return domain.Roster{
					ID:              rID,
					Name:            "Competition A",
					CountOfStudents: 1,
				}, nil
			},
		}

		reqBody := AddRosterStudentsRequest{
			StudentIds: []uuid.UUID{studentID},
		}
		bodyBytes, _ := json.Marshal(reqBody)
		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/rosters/"+rosterID.String()+"/students", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
	})

	t.Run("RemoveStudent_Unauthorized", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/"+rosterID.String()+"/students/"+studentID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("RemoveStudent_InvalidRosterID", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/bad-id/students/"+studentID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("RemoveStudent_InvalidStudentID", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/"+rosterID.String()+"/students/bad-id", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("RemoveStudent_ServiceError", func(t *testing.T) {
		mockSvc := &mockRosterService{
			removeStudentFn: func(ctx context.Context, rID, sID, ownerID uuid.UUID) error {
				return assert.AnError
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/"+rosterID.String()+"/students/"+studentID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})

	t.Run("RemoveStudent_Success", func(t *testing.T) {
		mockSvc := &mockRosterService{
			removeStudentFn: func(ctx context.Context, rID, sID, ownerID uuid.UUID) error {
				assert.Equal(t, rosterID, rID)
				assert.Equal(t, studentID, sID)
				return nil
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/rosters/"+rosterID.String()+"/students/"+studentID.String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNoContent, w.Code)
	})

	t.Run("UpdateStudentNote_Unauthorized", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String()+"/students/"+studentID.String(), bytes.NewBufferString(`{"note":"new"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("UpdateStudentNote_InvalidRosterID", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/bad-id/students/"+studentID.String(), bytes.NewBufferString(`{"note":"new"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("UpdateStudentNote_InvalidStudentID", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String()+"/students/bad-id", bytes.NewBufferString(`{"note":"new"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("UpdateStudentNote_InvalidJSON", func(t *testing.T) {
		r := setupRosterRouter(&mockRosterService{}, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String()+"/students/"+studentID.String(), bytes.NewBufferString(`{bad`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("UpdateStudentNote_ServiceError", func(t *testing.T) {
		mockSvc := &mockRosterService{
			updateParticipantNoteFn: func(ctx context.Context, rID, sID, ownerID uuid.UUID, note *string) error {
				return assert.AnError
			},
		}

		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String()+"/students/"+studentID.String(), bytes.NewBufferString(`{"note":"new"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})

	t.Run("UpdateStudentNote_Success", func(t *testing.T) {
		noteVal := "new note"
		mockSvc := &mockRosterService{
			updateParticipantNoteFn: func(ctx context.Context, rID, sID, ownerID uuid.UUID, note *string) error {
				assert.Equal(t, rosterID, rID)
				assert.Equal(t, studentID, sID)
				assert.Equal(t, &noteVal, note)
				return nil
			},
		}

		reqBody := UpdateRosterStudentNoteRequest{
			Note: &noteVal,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		r := setupRosterRouter(mockSvc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPatch, "/api/v1/rosters/"+rosterID.String()+"/students/"+studentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNoContent, w.Code)
	})
}
