package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/harness"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/db"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/model"
	"github.com/stretchr/testify/assert"
)

type mockUserService struct {
	deleteUserFn func(ctx context.Context, jwtUser authmodel.JwtUser) error
}

func (m *mockUserService) GetOrCreateUserFromOAuth2(ctx context.Context, provider db.UsersModuleOauth2Provider, providerID, email string) (response.UserResponse, error) {
	return response.UserResponse{}, nil
}

func (m *mockUserService) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	return true, nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, jwtUser authmodel.JwtUser) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, jwtUser)
	}
	return nil
}

func setupUserRouter(svc userService, user authmodel.JwtUser) *gin.Engine {
	return harness.SetupTestRouter(user, func(rg *gin.RouterGroup) {
		NewHandler(svc).RegisterRoutes(rg)
	})
}

func TestUserHandler_DeleteMyAccount(t *testing.T) {
	testUser := authmodel.JwtUser{ID: uuid.New(), Email: "user@example.com"}

	t.Run("Unauthorized", func(t *testing.T) {
		r := setupUserRouter(&mockUserService{}, authmodel.JwtUser{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/users/me", nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		deleteCalled := false
		svc := &mockUserService{
			deleteUserFn: func(ctx context.Context, jwtUser authmodel.JwtUser) error {
				assert.Equal(t, testUser.ID, jwtUser.ID)
				deleteCalled = true
				return nil
			},
		}

		r := setupUserRouter(svc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/users/me", nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNoContent, w.Code)
		assert.True(t, deleteCalled)
	})

	t.Run("NotFound", func(t *testing.T) {
		svc := &mockUserService{
			deleteUserFn: func(ctx context.Context, jwtUser authmodel.JwtUser) error {
				return pkgerrors.NewNotFound("OAuth2User", jwtUser.ID, jwtUser.ID)
			},
		}

		r := setupUserRouter(svc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/users/me", nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNotFound, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		svc := &mockUserService{
			deleteUserFn: func(ctx context.Context, jwtUser authmodel.JwtUser) error {
				return assert.AnError
			},
		}

		r := setupUserRouter(svc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/users/me", nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}
