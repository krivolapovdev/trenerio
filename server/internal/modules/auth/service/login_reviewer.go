package service

import (
	"context"

	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func (s *Service) LoginReviewer(ctx context.Context, reviewerKey string) (LoginResult, error) {
	logger.InfoContext(ctx, "Processing reviewer login")
	if !s.tokenProvider.ValidateReviewerKey(reviewerKey) {
		logger.WarnContext(ctx, "Invalid reviewer key provided")
		return LoginResult{}, pkgerrors.NewBadRequest("Invalid reviewer key")
	}

	u, err := s.userProvider.GetOrCreateUserFromOAuth2(ctx, db.UsersModuleOauth2ProviderGOOGLE, ReviewerProviderID, ReviewerEmail)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get or create user for reviewer login", "error", err)
		return LoginResult{}, err
	}

	resp, err := generateLoginResponse(ctx, s.tokenProvider, u, "reviewer")
	if err != nil {
		return LoginResult{}, err
	}
	logger.InfoContext(ctx, "Reviewer login successful", "user_id", u.ID.String())
	return resp, nil
}
