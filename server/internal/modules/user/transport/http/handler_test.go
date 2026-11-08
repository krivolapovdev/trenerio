package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/harness"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	usermodel "github.com/krivolapovdev/trenerio/internal/modules/user/domain"
	"github.com/stretchr/testify/assert"
)

type mockUserService struct {
	deleteUserFn func(ctx context.Context, jwtUser authmodel.JwtUser) error
}

func (m *mockUserService) GetOrCreateUserFromOAuth2() (usermodel.User, error) {
	return usermodel.User{}, nil
}

func (m *mockUserService) ExistsByID() (bool, error) {
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

	t.Run("ServiceError", func(t *testing.T) {
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

	t.Run("Success", func(t *testing.T) {
		svc := &mockUserService{
			deleteUserFn: func(ctx context.Context, jwtUser authmodel.JwtUser) error {
				assert.Equal(t, testUser.ID, jwtUser.ID)
				return nil
			},
		}
		r := setupUserRouter(svc, testUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/users/me", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusNoContent, w.Code)
	})
}
