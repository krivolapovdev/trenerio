package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
)

func (s *Service) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	logger.DebugContext(ctx, "Checking whether user exists", "id", id.String())

	exists, err := s.db.ExistsUserByID(ctx, id)

	if err != nil {
		logger.ErrorContext(ctx, "Failed to check if user exists", "id", id.String(), "error", err)
	}

	return exists, err
}
