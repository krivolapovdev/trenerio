package domain

import usermodel "github.com/krivolapovdev/trenerio/internal/modules/user/domain"

// @name Tokens
type Tokens struct {
	AccessToken  string
	RefreshToken string
}

// @name LoginResult
type LoginResult struct {
	User   usermodel.User
	Tokens Tokens
}
