package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCorsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CorsMiddleware([]string{"https://app.trenerio.ru"}))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	t.Run("OptionsPreflight", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
		req.Header.Set("Origin", "https://app.trenerio.ru")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "https://app.trenerio.ru", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	})

	t.Run("AllowedOrigin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Origin", "https://app.trenerio.ru")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://app.trenerio.ru", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("DisallowedOrigin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Origin", "https://malicious.org")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestCorsMiddleware_WildcardAndDisallowedPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("WildcardOriginReflectsAndAllowsCredentials", func(t *testing.T) {
		r := gin.New()
		r.Use(CorsMiddleware([]string{"*"}))
		r.GET("/api/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Origin", "https://my-frontend.domain.com")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://my-frontend.domain.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.NotContains(t, w.Header().Get("Access-Control-Expose-Headers"), "*")
	})

	t.Run("DisallowedOriginPreflightOptionsReturnsForbidden", func(t *testing.T) {
		r := gin.New()
		r.Use(CorsMiddleware([]string{"https://trusted.org"}))
		r.GET("/api/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
		req.Header.Set("Origin", "https://untrusted-hacker.com")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("EmptyAllowedOriginsAllowsAll", func(t *testing.T) {
		r := gin.New()
		r.Use(CorsMiddleware(nil))
		r.GET("/api/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Origin", "https://any-domain.com")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://any-domain.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("NoOriginHeader_PassesThrough", func(t *testing.T) {
		r := gin.New()
		r.Use(CorsMiddleware([]string{"https://app.trenerio.ru"}))
		r.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "pong", w.Body.String())
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("MultipleAllowedOrigins", func(t *testing.T) {
		allowed := []string{"https://app.trenerio.ru", "https://admin.trenerio.ru"}
		r := gin.New()
		r.Use(CorsMiddleware(allowed))
		r.GET("/data", func(c *gin.Context) {
			c.String(http.StatusOK, "data")
		})

		// First allowed origin
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/data", nil)
		req1.Header.Set("Origin", "https://app.trenerio.ru")
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Equal(t, "https://app.trenerio.ru", w1.Header().Get("Access-Control-Allow-Origin"))

		// Second allowed origin
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/data", nil)
		req2.Header.Set("Origin", "https://admin.trenerio.ru")
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Equal(t, "https://admin.trenerio.ru", w2.Header().Get("Access-Control-Allow-Origin"))

		// Third disallowed origin preflight
		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodOptions, "/data", nil)
		req3.Header.Set("Origin", "https://other.com")
		r.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusForbidden, w3.Code)
	})
}
