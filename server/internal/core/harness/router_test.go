package harness

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/stretchr/testify/assert"
)

func TestSetupTestRouter(t *testing.T) {
	testUser := authmodel.JwtUser{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	t.Run("AuthenticatedUser", func(t *testing.T) {
		r := SetupTestRouter(testUser, func(rg *gin.RouterGroup) {
			rg.GET("/ping", func(c *gin.Context) {
				c.String(http.StatusOK, "pong")
			})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "pong", w.Body.String())
	})

	t.Run("AnonymousUser", func(t *testing.T) {
		r := SetupTestRouter(authmodel.JwtUser{}, func(rg *gin.RouterGroup) {
			rg.GET("/ping", func(c *gin.Context) {
				c.String(http.StatusOK, "anonymous")
			})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "anonymous", w.Body.String())
	})

	t.Run("ContextRetrievalInHandler", func(t *testing.T) {
		var retrievedUser authmodel.JwtUser
		var retrievedOk bool
		r := SetupTestRouter(testUser, func(rg *gin.RouterGroup) {
			rg.GET("/user", func(c *gin.Context) {
				retrievedUser, retrievedOk = authhttp.GetUserFromContext(c)
				c.Status(http.StatusOK)
			})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, retrievedOk)
		assert.Equal(t, testUser.ID, retrievedUser.ID)
		assert.Equal(t, testUser.Email, retrievedUser.Email)
	})
}
