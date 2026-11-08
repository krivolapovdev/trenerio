package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authmodel "github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.Use(LoggingMiddleware())
	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})
	r.GET("/client-err", func(c *gin.Context) {
		c.String(http.StatusBadRequest, "bad input")
	})
	r.GET("/server-err", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "error")
	})
	r.GET("/with-user-and-query", func(c *gin.Context) {
		c.Set(authhttp.ContextKeyJWTUser, authmodel.JwtUser{ID: uuid.New(), Email: "coach@trenerio.ru"})
		_ = c.Error(errors.New("attached gin error"))
		c.String(http.StatusOK, "ok")
	})

	t.Run("SuccessLogging", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ClientErrorLogging", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/client-err", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServerErrorLogging", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/server-err", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("WithUserQueryAndGinErrors", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/with-user-and-query?foo=bar", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("LoggingWithoutRequestIDMiddleware", func(t *testing.T) {
		r2 := gin.New()
		r2.Use(LoggingMiddleware())
		r2.GET("/test-header-req-id", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		// 1. With X-Request-ID header
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test-header-req-id", nil)
		req.Header.Set(web.HeaderXRequestID, "header-req-id")
		r2.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 2. Without X-Request-ID header (auto generates uuid)
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/test-header-req-id", nil)
		r2.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
}
