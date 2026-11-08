package http

import (
	"bytes"
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	loginGoogleFn   func(ctx context.Context, token string) (response.LoginResponse, error)
	refreshTokensFn func(ctx context.Context, refreshToken string) (response.JwtResponse, error)
	loginReviewerFn func(ctx context.Context, reviewerKey string) (response.LoginResponse, error)
}

func (m *mockAuthService) LoginGoogle(ctx context.Context, token string) (response.LoginResponse, error) {
	if m.loginGoogleFn != nil {
		return m.loginGoogleFn(ctx, token)
	}
	return response.LoginResponse{}, nil
}

func (m *mockAuthService) RefreshTokens(ctx context.Context, refreshToken string) (response.JwtResponse, error) {
	if m.refreshTokensFn != nil {
		return m.refreshTokensFn(ctx, refreshToken)
	}
	return response.JwtResponse{}, nil
}

func (m *mockAuthService) LoginReviewer(ctx context.Context, reviewerKey string) (response.LoginResponse, error) {
	if m.loginReviewerFn != nil {
		return m.loginReviewerFn(ctx, reviewerKey)
	}
	return response.LoginResponse{}, nil
}

func setupAuthRouter(svc AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestAuthHandler_GoogleLogin(t *testing.T) {
	testUser := response.UserResponse{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	testTokens := response.JwtResponse{
		AccessToken:  "valid-access-token",
		RefreshToken: "valid-refresh-token",
	}

	t.Run("Success", func(t *testing.T) {
		svc := &mockAuthService{
			loginGoogleFn: func(ctx context.Context, token string) (response.LoginResponse, error) {
				assert.Equal(t, "google-id-token", token)
				return response.LoginResponse{
					User:      testUser,
					JwtTokens: testTokens,
				}, nil
			},
		}

		r := setupAuthRouter(svc)
		body, _ := json.Marshal(map[string]string{"token": "google-id-token"})
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/oauth2/google", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
		var resp response.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, resp.User.ID)
		assert.Equal(t, "valid-access-token", resp.JwtTokens.AccessToken)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		svc := &mockAuthService{}
		r := setupAuthRouter(svc)

		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/oauth2/google", bytes.NewBufferString("{invalid-json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		svc := &mockAuthService{
			loginGoogleFn: func(ctx context.Context, token string) (response.LoginResponse, error) {
				return response.LoginResponse{}, pkgerrors.NewBadRequest("invalid token")
			},
		}

		r := setupAuthRouter(svc)
		body, _ := json.Marshal(map[string]string{"token": "bad-token"})
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/oauth2/google", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_ReviewerLogin(t *testing.T) {
	t.Run("MissingHeader", func(t *testing.T) {
		mockSvc := &mockAuthService{}
		r := setupAuthRouter(mockSvc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/reviewer/login", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		expectedResp := response.LoginResponse{
			User: response.UserResponse{
				ID:    uuid.New(),
				Email: "REVIEWER@TRENERIO.RU",
			},
		}

		mockSvc := &mockAuthService{
			loginReviewerFn: func(ctx context.Context, reviewerKey string) (response.LoginResponse, error) {
				assert.Equal(t, "secret-key", reviewerKey)
				return expectedResp, nil
			},
		}

		r := setupAuthRouter(mockSvc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/reviewer/login", nil)
		req.Header.Set(web.HeaderXReviewerKey, "secret-key")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
	})

	t.Run("Unauthorized_InvalidKey", func(t *testing.T) {
		mockSvc := &mockAuthService{
			loginReviewerFn: func(ctx context.Context, reviewerKey string) (response.LoginResponse, error) {
				return response.LoginResponse{}, pkgerrors.NewUnauthorized("Invalid reviewer key")
			},
		}

		r := setupAuthRouter(mockSvc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/reviewer/login", nil)
		req.Header.Set(web.HeaderXReviewerKey, "wrong-key")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusUnauthorized, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockSvc := &mockAuthService{
			loginReviewerFn: func(ctx context.Context, reviewerKey string) (response.LoginResponse, error) {
				return response.LoginResponse{}, assert.AnError
			},
		}

		r := setupAuthRouter(mockSvc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/reviewer/login", nil)
		req.Header.Set(web.HeaderXReviewerKey, "crash-key")
		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusInternalServerError, w.Code)
	})
}

func TestAuthHandler_RefreshTokens(t *testing.T) {
	testTokens := response.JwtResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	t.Run("Success", func(t *testing.T) {
		svc := &mockAuthService{
			refreshTokensFn: func(ctx context.Context, refreshToken string) (response.JwtResponse, error) {
				assert.Equal(t, "valid-refresh-token", refreshToken)
				return testTokens, nil
			},
		}

		r := setupAuthRouter(svc)
		body, _ := json.Marshal(map[string]string{"refreshToken": "valid-refresh-token"})
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/jwt/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusOK, w.Code)
		var resp response.JwtResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "new-access-token", resp.AccessToken)
		assert.Equal(t, "new-refresh-token", resp.RefreshToken)
	})

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		svc := &mockAuthService{
			refreshTokensFn: func(ctx context.Context, refreshToken string) (response.JwtResponse, error) {
				return response.JwtResponse{}, pkgerrors.NewBadRequest("Invalid or expired refresh token")
			},
		}

		r := setupAuthRouter(svc)
		body, _ := json.Marshal(map[string]string{"refreshToken": "bad-token"})
		req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/jwt/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, nethttp.StatusBadRequest, w.Code)
	})
}
