package mapper

import (
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/db"
)

func ToUserResponse(u db.UsersModuleOauth2User) response.UserResponse {
	return response.UserResponse{
		ID:    u.ID,
		Email: u.Email,
	}
}
