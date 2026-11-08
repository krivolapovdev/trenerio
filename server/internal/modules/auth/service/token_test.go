package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserProvider struct {
	getOrCreateUserFromOAuth2 func(ctx context.Context, provider db.UsersModuleOauth2Provider, providerID, email string) (UserResponse, error)
	existsByID                func(ctx context.Context, id uuid.UUID) (bool, error)
}

func (m *mockUserProvider) GetOrCreateUserFromOAuth2(ctx context.Context, provider db.UsersModuleOauth2Provider, providerID, email string) (UserResponse, error) {
	if m.getOrCreateUserFromOAuth2 != nil {
		return m.getOrCreateUserFromOAuth2(ctx, provider, providerID, email)
	}
	return UserResponse{}, nil
}

func (m *mockUserProvider) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.existsByID != nil {
		return m.existsByID(ctx, id)
	}
	return false, nil
}

func TestAuthService_LoginReviewer(t *testing.T) {
	reviewerKey := "super-secret-reviewer-key"
	tokenProvider := NewTokenProvider("test-jwt-secret-key-at-least-32-chars", time.Hour, 24*time.Hour, reviewerKey)
	testUserID := uuid.New()

	mockProvider := &mockUserProvider{
		getOrCreateUserFromOAuth2: func(ctx context.Context, provider db.UsersModuleOauth2Provider, providerID, email string) (UserResponse, error) {
			assert.Equal(t, ReviewerProviderID, providerID)
			assert.Equal(t, ReviewerEmail, email)
			return UserResponse{
				ID:    testUserID,
				Email: email,
			}, nil
		},
	}

	service := NewService(mockProvider, tokenProvider, "dummy-google-client-id")

	t.Run("ValidKey", func(t *testing.T) {
		resp, err := service.LoginReviewer(context.Background(), reviewerKey)
		require.NoError(t, err)
		assert.Equal(t, testUserID, resp.User.ID)
		assert.Equal(t, ReviewerEmail, resp.User.Email)
		assert.NotEmpty(t, resp.JwtTokens.AccessToken)
		assert.NotEmpty(t, resp.JwtTokens.RefreshToken)
	})

	t.Run("InvalidKey", func(t *testing.T) {
		_, err := service.LoginReviewer(context.Background(), "invalid-key")
		require.Error(t, err)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
	})
}

func TestAuthService_RefreshTokens(t *testing.T) {
	tokenProvider := NewTokenProvider("test-jwt-secret-key-at-least-32-chars", time.Hour, 24*time.Hour, "reviewer-key")
	testUserID := uuid.New()

	validUser := JwtUser{
		ID:    testUserID,
		Email: "refresh@example.com",
	}

	validRefreshToken, err := tokenProvider.GenerateRefreshToken(validUser)
	require.NoError(t, err)

	validAccessToken, err := tokenProvider.GenerateAccessToken(validUser)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		mockProvider := &mockUserProvider{
			existsByID: func(ctx context.Context, id uuid.UUID) (bool, error) {
				return id == testUserID, nil
			},
		}

		service := NewService(mockProvider, tokenProvider, "dummy-google-client-id")
		tokens, err := service.RefreshTokens(context.Background(), validRefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	t.Run("InvalidTokenType", func(t *testing.T) {
		mockProvider := &mockUserProvider{}
		service := NewService(mockProvider, tokenProvider, "dummy-google-client-id")
		_, err := service.RefreshTokens(context.Background(), validAccessToken)
		require.Error(t, err)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockProvider := &mockUserProvider{
			existsByID: func(ctx context.Context, id uuid.UUID) (bool, error) {
				return false, nil
			},
		}

		service := NewService(mockProvider, tokenProvider, "dummy-google-client-id")
		_, err := service.RefreshTokens(context.Background(), validRefreshToken)
		require.Error(t, err)
		var badReq *pkgerrors.BadRequestError
		require.ErrorAs(t, err, &badReq)
	})
}
