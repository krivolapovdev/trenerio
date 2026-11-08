package service

import (
	"context"

	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"google.golang.org/api/idtoken"
)

func (s *Service) LoginGoogle(ctx context.Context, token string) (LoginResult, error) {
	logger.InfoContext(ctx, "Processing Google login")
	payload, err := idtoken.Validate(ctx, token, s.googleClientID)
	if err != nil {
		logger.WarnContext(ctx, "Failed to validate Google identity token", "error", err)
		return LoginResult{}, pkgerrors.NewBadRequest("Invalid or expired Google identity token")
	}

	providerID := payload.Subject
	if providerID == "" {
		logger.WarnContext(ctx, "Google identity token does not contain a subject")
		return LoginResult{}, pkgerrors.NewBadRequest("Google identity token does not contain a subject")
	}

	emailVal, ok := payload.Claims["email"]
	if !ok {
		logger.WarnContext(ctx, "Google identity token missing email claim", "subject", providerID)
		return LoginResult{}, pkgerrors.NewBadRequest("Google identity token does not contain an email")
	}
	email, ok := emailVal.(string)
	if !ok || email == "" {
		logger.WarnContext(ctx, "Google identity token has invalid email claim", "subject", providerID)
		return LoginResult{}, pkgerrors.NewBadRequest("Google identity token does not contain an email")
	}

	logger.DebugContext(ctx, "Valid Google token received", "provider_id", providerID, "email", email)

	u, err := s.userProvider.GetOrCreateUserFromOAuth2(ctx, db.UsersModuleOauth2ProviderGOOGLE, providerID, email)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get or create user for Google login", "error", err, "email", email)
		return LoginResult{}, err
	}

	resp, err := generateLoginResponse(ctx, s.tokenProvider, u, "Google")
	if err != nil {
		return LoginResult{}, err
	}
	logger.InfoContext(ctx, "Google login successful", "user_id", u.ID.String(), "email", u.Email)
	return resp, nil
}
