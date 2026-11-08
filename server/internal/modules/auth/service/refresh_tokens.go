package service

import (
	"context"

	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
)

func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error) {
	logger.InfoContext(ctx, "Refreshing tokens")
	if !s.tokenProvider.IsTokenValid(refreshToken, TokenTypeRefresh) {
		logger.WarnContext(ctx, "Invalid or expired refresh token provided")
		return Tokens{}, pkgerrors.NewBadRequest("Invalid or expired refresh token")
	}

	jwtUser, err := s.tokenProvider.ExtractUser(refreshToken)
	if err != nil {
		logger.WarnContext(ctx, "Failed to extract user from refresh token", "error", err)
		return Tokens{}, pkgerrors.NewBadRequest("Invalid or expired refresh token")
	}

	logger.DebugContext(ctx, "Token user extracted", "user_id", jwtUser.ID.String(), "email", jwtUser.Email)

	exists, err := s.userProvider.ExistsByID(ctx, jwtUser.ID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to check user existence during token refresh", "user_id", jwtUser.ID.String(), "error", err)
		return Tokens{}, err
	}
	if !exists {
		logger.WarnContext(ctx, "User not found during token refresh", "user_id", jwtUser.ID.String())
		return Tokens{}, pkgerrors.NewBadRequest("Invalid or expired refresh token")
	}

	accessToken, err := s.tokenProvider.GenerateAccessToken(jwtUser)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate access token during refresh", "user_id", jwtUser.ID.String(), "error", err)
		return Tokens{}, err
	}
	newRefreshToken, err := s.tokenProvider.GenerateRefreshToken(jwtUser)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate refresh token during refresh", "user_id", jwtUser.ID.String(), "error", err)
		return Tokens{}, err
	}

	logger.InfoContext(ctx, "Tokens refreshed successfully", "user_id", jwtUser.ID.String())
	return Tokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
