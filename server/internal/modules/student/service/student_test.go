package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/types"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStudentService_DateAndPaymentValidation(t *testing.T) {
	svc := NewService(nil, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	t.Run("CreateStudent_FutureBirthdate", func(t *testing.T) {
		futureBirthdate := types.Date{Time: time.Now().UTC().Add(24 * time.Hour)}
		_, err := svc.CreateStudent(ctx, studentmodel.Student{
			FullName:  "Test Student",
			Birthdate: &futureBirthdate,
		}, ownerID)
		require.Error(t, err)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Equal(t, "birthdate cannot be in the future", badReq.Message)
	})

	t.Run("CreateStudentPayment_FutureDate", func(t *testing.T) {
		futureDate := time.Now().UTC().Add(24 * time.Hour)
		paidUntil := futureDate.Add(24 * time.Hour)
		_, err := svc.CreateStudentPayment(ctx, uuid.New(), ownerID, 1000, futureDate, nil, &paidUntil, nil)
		require.Error(t, err)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Equal(t, "date cannot be in the future", badReq.Message)
	})

	t.Run("CreateStudentPayment_MissingPaidUntilAndSessionsAdded", func(t *testing.T) {
		pastDate := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
		_, err := svc.CreateStudentPayment(ctx, uuid.New(), ownerID, 1000, pastDate, nil, nil, nil)
		require.Error(t, err)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
		assert.Contains(t, badReq.Message, "either paidUntil or sessionsAdded")
	})

	t.Run("GetStudentsByIDs_Empty", func(t *testing.T) {
		res, err := svc.GetStudentsByIDs(ctx, []uuid.UUID{}, ownerID)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestStudentService_Dependencies(t *testing.T) {
	svc := NewService(nil, nil)
	require.NotNil(t, svc)

	svc.SetDependencies(nil, nil, nil, nil)
	assert.Nil(t, svc.visitProvider)
	assert.Nil(t, svc.transactionProvider)
	assert.Nil(t, svc.groupProvider)
	assert.Nil(t, svc.lessonProvider)
}
