package harness

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/model"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

func SetupTestRouter(user authmodel.JwtUser, registerRoutes func(rg *gin.RouterGroup)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if user.ID != uuid.Nil {
			c.Set(authhttp.ContextKeyJWTUser, user)
		}
		c.Next()
	})
	registerRoutes(r.Group("/api/v1"))
	return r
}
