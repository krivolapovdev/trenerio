package http

import (
	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/modules/auth/model"
)

type JwtUser = model.JwtUser

const (
	ContextKeyJWTUser = "jwtUser"
)

func CurrentUser(c *gin.Context) (JwtUser, bool) {
	return GetUserFromContext(c)
}

func SetUserToContext(c *gin.Context, u JwtUser) {
	c.Set(ContextKeyJWTUser, u)
}

func GetUserFromContext(c *gin.Context) (JwtUser, bool) {
	val, exists := c.Get(ContextKeyJWTUser)
	if !exists {
		pkgerrors.RespondUnauthorized(c, "")
		c.Abort()
		return JwtUser{}, false
	}

	u, ok := val.(JwtUser)
	if !ok {
		pkgerrors.RespondUnauthorized(c, "")
		c.Abort()
		return JwtUser{}, false
	}

	return u, true
}
