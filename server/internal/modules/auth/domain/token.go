package domain

import usermodel "github.com/krivolapovdev/trenerio/internal/modules/user/domain"

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type LoginResult struct {
	User   usermodel.User
	Tokens Tokens
}
