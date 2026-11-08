package mapper

import (
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/user/domain"
)

func ToUserModel(u db.UsersModuleOauth2User) domain.User {
	return domain.User{
		ID:    u.ID,
		Email: u.Email,
	}
}
