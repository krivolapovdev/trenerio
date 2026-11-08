package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/modules/roster/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStudentProvider struct {
	getStudentsByIDsFn func(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) ([]studentmodel.Student, error)
}

func (m *mockStudentProvider) GetStudentsByIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) ([]studentmodel.Student, error) {
	if m.getStudentsByIDsFn != nil {
		return m.getStudentsByIDsFn(ctx, studentIDs, ownerID)
	}
	return nil, nil
}

func TestRosterService_Validation(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("CreateRoster_BlankName", func(t *testing.T) {
		svc := NewService(nil, nil, nil)
		_, err := svc.CreateRoster(ctx, domain.Roster{
			Name: "   ",
		}, nil, ownerID)
		require.Error(t, err)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Equal(t, "name must not be blank", badReq.Message)
	})

	t.Run("CreateRoster_StudentNotFound", func(t *testing.T) {
		missingID := uuid.New()
		studentProvider := &mockStudentProvider{
			getStudentsByIDsFn: func(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) ([]studentmodel.Student, error) {
				return []studentmodel.Student{}, nil // returns 0 students found
			},
		}

		svc := NewService(nil, nil, studentProvider)
		_, err := svc.CreateRoster(ctx, domain.Roster{
			Name: "Competition A",
		}, []uuid.UUID{missingID}, ownerID)
		require.Error(t, err)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Equal(t, "One or more student IDs not found", badReq.Message)
	})
}
