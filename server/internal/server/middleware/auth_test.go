package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	authsvc "github.com/krivolapovdev/trenerio/internal/modules/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenProvider := authsvc.NewTokenProvider("super-secret-key-32-bytes-long!", time.Hour, 24*time.Hour, "reviewer-key")
	user := authmodel.JwtUser{
		ID:    uuid.New(),
		Email: "teacher@trenerio.ru",
	}

	validAccessToken, err := tokenProvider.GenerateAccessToken(user)
	require.NoError(t, err)

	validRefreshToken, err := tokenProvider.GenerateRefreshToken(user)
	require.NoError(t, err)

	r := gin.New()
	r.Use(AuthMiddleware(tokenProvider))
	r.GET("/protected", func(c *gin.Context) {
		val, exists := c.Get("jwtUser")
		if !exists {
			c.Status(http.StatusInternalServerError)
			return
		}
		u := val.(authmodel.JwtUser)
		c.JSON(http.StatusOK, gin.H{
			"id":    u.ID.String(),
			"email": u.Email,
		})
	})

	t.Run("EmptyBearerToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer    ")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ValidTokenWithInvalidSubjectUUID", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":   "not-a-valid-uuid",
			"type":  "ACCESS",
			"email": "teacher@trenerio.ru",
			"exp":   time.Now().Add(time.Hour).Unix(),
		}
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := rawToken.SignedString([]byte("super-secret-key-32-bytes-long!"))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("MissingAuthorizationHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var pd pkgerrors.ProblemDetail
		err := json.Unmarshal(w.Body.Bytes(), &pd)
		require.NoError(t, err)
		assert.Equal(t, "Unauthorized", pd.Title)
		assert.Equal(t, "Full authentication is required to access this resource", pd.Detail)
	})

	t.Run("InvalidBearerToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("RefreshTokenUsedAsBearer", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validRefreshToken)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ValidAccessToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validAccessToken)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var res map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, user.ID.String(), res["id"])
		assert.Equal(t, user.Email, res["email"])
	})
}
