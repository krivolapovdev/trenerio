package service

import (
	"context"

	"github.com/krivolapovdev/trenerio/internal/core/logger"
	usermodel "github.com/krivolapovdev/trenerio/internal/modules/user/domain"
)

func generateLoginResponse(ctx context.Context, tp *TokenProvider, u usermodel.User, loginType string) (LoginResult, error) {
	jwtUser := JwtUser{
		ID:    u.ID,
		Email: u.Email,
	}

	accessToken, refreshToken, err := tp.GenerateTokenPair(jwtUser)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to generate tokens for "+loginType+" login", "error", err, "user_id", u.ID.String())
		return LoginResult{}, err
	}

	return LoginResult{
		User: u,
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}
