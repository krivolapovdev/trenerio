package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/user/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/user/mapper"
)

func (s *Service) GetOrCreateUserFromOAuth2(ctx context.Context, provider db.UsersModuleOauth2Provider, providerID, email string) (domain.User, error) {
	logger.InfoContext(ctx, "Get or create user from OAuth2", "provider", string(provider), "email", email)

	existingUser, err := s.db.GetUserByProviderAndProviderID(ctx, db.GetUserByProviderAndProviderIDParams{
		Provider:   provider,
		ProviderID: providerID,
	})

	if err == nil {
		logger.DebugContext(ctx, "Existing OAuth2 user found", "user_id", existingUser.ID.String(), "provider", string(provider))
		return mapper.ToUserModel(existingUser), nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		logger.ErrorContext(ctx, "Failed to query OAuth2 user", "error", err, "provider", string(provider))
		return domain.User{}, err
	}

	normalizedEmail := strings.ToUpper(strings.TrimSpace(email))

	logger.InfoContext(ctx, "Creating new user", "provider", string(provider), "email", normalizedEmail)

	created, err := s.db.CreateUser(ctx, db.CreateUserParams{
		Provider:   provider,
		ProviderID: providerID,
		Email:      normalizedEmail,
	})

	if err != nil {
		logger.ErrorContext(ctx, "Failed to create user in database", "error", err, "provider", string(provider), "email", normalizedEmail)
		return domain.User{}, err
	}

	logger.InfoContext(ctx, "User created successfully", "user_id", created.ID.String(), "email", created.Email)

	return mapper.ToUserModel(created), nil
}
