package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandlerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("HandlesGinError", func(t *testing.T) {
		r := gin.New()
		r.Use(ErrorHandlerMiddleware())
		r.GET("/err", func(c *gin.Context) {
			_ = c.Error(errors.New("some unexpected database error"))
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/err", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal Server Error")
	})

	t.Run("PassesThroughWhenNoError", func(t *testing.T) {
		r := gin.New()
		r.Use(ErrorHandlerMiddleware())
		r.GET("/ok", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("SkipsIfAlreadyWritten", func(t *testing.T) {
		r := gin.New()
		r.Use(ErrorHandlerMiddleware())
		r.GET("/written", func(c *gin.Context) {
			c.String(http.StatusAccepted, "custom response")
			_ = c.Error(errors.New("ignored error after write"))
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/written", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Equal(t, "custom response", w.Body.String())
	})

	t.Run("HandlesTypedErrors", func(t *testing.T) {
		r := gin.New()
		r.Use(ErrorHandlerMiddleware())
		r.GET("/not-found", func(c *gin.Context) {
			_ = c.Error(pkgerrors.NewNotFound("Group", uuid.New(), uuid.New()))
		})
		r.GET("/bad-request", func(c *gin.Context) {
			_ = c.Error(pkgerrors.NewBadRequest("invalid field"))
		})
		r.GET("/unauthorized", func(c *gin.Context) {
			_ = c.Error(pkgerrors.NewUnauthorized("access denied"))
		})
		r.GET("/conflict", func(c *gin.Context) {
			_ = c.Error(pkgerrors.NewConflict("CONFLICT_CODE", "already exists"))
		})
		r.GET("/validation", func(c *gin.Context) {
			_ = c.Error(pkgerrors.NewValidationError([]map[string]any{{"field": "name", "error": "required"}}))
		})

		// 404
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/not-found", nil)
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusNotFound, w1.Code)

		// 400
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/bad-request", nil)
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusBadRequest, w2.Code)

		// 401
		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodGet, "/unauthorized", nil)
		r.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusUnauthorized, w3.Code)

		// 409
		w4 := httptest.NewRecorder()
		req4 := httptest.NewRequest(http.MethodGet, "/conflict", nil)
		r.ServeHTTP(w4, req4)
		assert.Equal(t, http.StatusConflict, w4.Code)

		// 400 validation
		w5 := httptest.NewRecorder()
		req5 := httptest.NewRequest(http.MethodGet, "/validation", nil)
		r.ServeHTTP(w5, req5)
		assert.Equal(t, http.StatusBadRequest, w5.Code)
	})

	t.Run("HandlesMultipleErrors_UsesLastError", func(t *testing.T) {
		r := gin.New()
		r.Use(ErrorHandlerMiddleware())
		r.GET("/multi-err", func(c *gin.Context) {
			_ = c.Error(errors.New("first error"))
			_ = c.Error(pkgerrors.NewBadRequest("last error is bad request"))
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/multi-err", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "last error is bad request")
	})
}
