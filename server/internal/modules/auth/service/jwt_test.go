package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenProvider_AccessAndRefreshToken(t *testing.T) {
	secret := "test-super-secret-key-that-is-long-enough-32-bytes"
	reviewerKey := "valid-reviewer-secret"
	provider := NewTokenProvider(secret, time.Hour, 24*time.Hour, reviewerKey)

	user := JwtUser{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	accessToken, err := provider.GenerateAccessToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	refreshToken, err := provider.GenerateRefreshToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	assert.True(t, provider.IsTokenValid(accessToken, "ACCESS"))
	assert.False(t, provider.IsTokenValid(accessToken, "REFRESH"))
	assert.True(t, provider.IsTokenValid(refreshToken, "REFRESH"))
	assert.False(t, provider.IsTokenValid(refreshToken, "ACCESS"))

	extracted, err := provider.ExtractUser(accessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, extracted.ID)
	assert.Equal(t, user.Email, extracted.Email)

	extractedRefresh, err := provider.ExtractUser(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, extractedRefresh.ID)
	assert.Equal(t, user.Email, extractedRefresh.Email)
}

func TestTokenProvider_InvalidTokens(t *testing.T) {
	provider := NewTokenProvider("key1-super-secret-key-32-bytes-long", time.Hour, time.Hour, "rev")
	otherProvider := NewTokenProvider("key2-super-secret-key-32-bytes-long", time.Hour, time.Hour, "rev")

	user := JwtUser{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	token, err := provider.GenerateAccessToken(user)
	require.NoError(t, err)

	assert.False(t, otherProvider.IsTokenValid(token, "ACCESS"), "Token signed by another provider must be invalid")

	_, err = otherProvider.ExtractUser(token)
	assert.Error(t, err)

	assert.False(t, provider.IsTokenValid("malformed.jwt.token", "ACCESS"))
	_, err = provider.ExtractUser("malformed.jwt.token")
	assert.Error(t, err)
}

func TestTokenProvider_ValidateReviewerKey(t *testing.T) {
	provider := NewTokenProvider("key-super-secret-key-32-bytes-long", time.Hour, time.Hour, "correct-key")

	assert.True(t, provider.ValidateReviewerKey("correct-key"))
	assert.False(t, provider.ValidateReviewerKey("wrong-key"))
	assert.False(t, provider.ValidateReviewerKey(""))

	emptyProvider := NewTokenProvider("key-super-secret-key-32-bytes-long", time.Hour, time.Hour, "")
	assert.False(t, emptyProvider.ValidateReviewerKey("correct-key"))
	assert.False(t, emptyProvider.ValidateReviewerKey(""))
}

func TestTokenProvider_DefaultExpirations(t *testing.T) {
	provider := NewTokenProvider("key-super-secret-key-32-bytes-long", 0, 0, "rev")
	assert.Equal(t, 15*time.Minute, provider.accessTokenExpiration)
	assert.Equal(t, 14*24*time.Hour, provider.refreshTokenExpiration)
}

func TestTokenProvider_GenerateTokenPair(t *testing.T) {
	provider := NewTokenProvider("key-super-secret-key-32-bytes-long", time.Hour, time.Hour, "rev")
	user := JwtUser{
		ID:    uuid.New(),
		Email: "pair@example.com",
	}

	access, refresh, err := provider.GenerateTokenPair(user)
	require.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	assert.True(t, provider.IsTokenValid(access, TokenTypeAccess))
	assert.True(t, provider.IsTokenValid(refresh, TokenTypeRefresh))
}
