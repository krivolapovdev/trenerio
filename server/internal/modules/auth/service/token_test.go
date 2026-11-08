package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/db"
	usermodel "github.com/krivolapovdev/trenerio/internal/modules/user/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserProvider struct {
	getOrCreateUserFromOAuth2 func(ctx context.Context, provider db.UsersModuleOauth2Provider, providerID, email string) (usermodel.User, error)
	existsByID                func(ctx context.Context, id uuid.UUID) (bool, error)
}

func (m *mockUserProvider) GetOrCreateUserFromOAuth2(ctx context.Context, provider db.UsersModuleOauth2Provider, providerID, email string) (usermodel.User, error) {
	if m.getOrCreateUserFromOAuth2 != nil {
		return m.getOrCreateUserFromOAuth2(ctx, provider, providerID, email)
	}
	return usermodel.User{}, nil
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
		getOrCreateUserFromOAuth2: func(ctx context.Context, provider db.UsersModuleOauth2Provider, providerID, email string) (usermodel.User, error) {
			assert.Equal(t, ReviewerProviderID, providerID)
			assert.Equal(t, ReviewerEmail, email)
			return usermodel.User{
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
		assert.NotEmpty(t, resp.Tokens.AccessToken)
		assert.NotEmpty(t, resp.Tokens.RefreshToken)
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
	testUser := JwtUser{
		ID:    uuid.New(),
		Email: "refresh@example.com",
	}

	validRefreshToken, err := tokenProvider.GenerateRefreshToken(testUser)
	require.NoError(t, err)

	mockProvider := &mockUserProvider{
		existsByID: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return id == testUser.ID, nil
		},
	}

	service := NewService(mockProvider, tokenProvider, "dummy-google-client-id")

	t.Run("ValidRefreshToken", func(t *testing.T) {
		tokens, err := service.RefreshTokens(context.Background(), validRefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		_, err := service.RefreshTokens(context.Background(), "totally-invalid-token")
		require.Error(t, err)
	})

	t.Run("InvalidTokenType", func(t *testing.T) {
		accessToken, err := tokenProvider.GenerateAccessToken(testUser)
		require.NoError(t, err)

		_, err = service.RefreshTokens(context.Background(), accessToken)
		require.Error(t, err)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		nonExistentUser := JwtUser{
			ID:    uuid.New(),
			Email: "ghost@example.com",
		}
		ghostToken, err := tokenProvider.GenerateRefreshToken(nonExistentUser)
		require.NoError(t, err)

		_, err = service.RefreshTokens(context.Background(), ghostToken)
		require.Error(t, err)
	})
}
