package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	usermodel "github.com/krivolapovdev/trenerio/internal/modules/user/domain"
)

const (
	ReviewerProviderID = "REVIEWER"
	ReviewerEmail      = "REVIEWER@TRENERIO.RU"
)

type UserProvider interface {
	GetOrCreateUserFromOAuth2(ctx context.Context, provider db.UsersModuleOauth2Provider, providerID, email string) (usermodel.User, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

type Service struct {
	userProvider   UserProvider
	tokenProvider  *TokenProvider
	googleClientID string
}

func NewService(
	userProvider UserProvider,
	tokenProvider *TokenProvider,
	googleClientID string,
) *Service {
	return &Service{
		userProvider:   userProvider,
		tokenProvider:  tokenProvider,
		googleClientID: googleClientID,
	}
}
