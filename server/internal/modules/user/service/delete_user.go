package service

import (
	"context"

	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
)

func (s *Service) DeleteUser(ctx context.Context, jwtUser authmodel.JwtUser) error {
	logger.InfoContext(ctx, "Deleting user", "user_id", jwtUser.ID.String())

	exists, err := s.db.ExistsUserByID(ctx, jwtUser.ID)

	if err != nil {
		logger.ErrorContext(ctx, "Failed to check user existence for deletion", "user_id", jwtUser.ID.String(), "error", err)
		return err
	}

	if !exists {
		logger.WarnContext(ctx, "User to delete not found", "user_id", jwtUser.ID.String())
		return pkgerrors.NewNotFound("OAuth2User", jwtUser.ID, jwtUser.ID)
	}

	if err := s.db.DeleteUserByID(ctx, jwtUser.ID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete user from database", "user_id", jwtUser.ID.String(), "error", err)
		return err
	}

	logger.InfoContext(ctx, "User deleted successfully", "user_id", jwtUser.ID.String())

	return nil
}
