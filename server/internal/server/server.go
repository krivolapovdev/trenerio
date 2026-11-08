package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	authsvc "github.com/krivolapovdev/trenerio/internal/modules/auth/service"
	"github.com/krivolapovdev/trenerio/internal/server/middleware"
	"github.com/krivolapovdev/trenerio/internal/server/swagger"
)

// RouteRegistrar defines a component capable of registering its routes onto a gin router group.
type RouteRegistrar interface {
	RegisterRoutes(rg *gin.RouterGroup)
}

type Server struct {
	engine        *gin.Engine
	tokenProvider *authsvc.TokenProvider

	authRouter       RouteRegistrar
	protectedRouters []RouteRegistrar
}

func NewServer(
	tokenProvider *authsvc.TokenProvider,
	allowedOrigins []string,
	authRouter RouteRegistrar,
	protectedRouters ...RouteRegistrar,
) *Server {
	engine := gin.New()
	engine.HandleMethodNotAllowed = true

	err := engine.SetTrustedProxies([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	})
	if err != nil {
		panic(err)
	}

	s := &Server{
		engine:           engine,
		tokenProvider:    tokenProvider,
		authRouter:       authRouter,
		protectedRouters: protectedRouters,
	}

	s.setupRoutes(allowedOrigins)

	return s
}

func (s *Server) setupRoutes(allowedOrigins []string) {
	s.engine.Use(middleware.RequestIDMiddleware())
	s.engine.Use(middleware.LoggingMiddleware())
	s.engine.Use(middleware.RecoveryMiddleware())
	s.engine.Use(middleware.CorsMiddleware(allowedOrigins))
	s.engine.Use(middleware.ErrorHandlerMiddleware())

	s.engine.NoRoute(func(c *gin.Context) {
		pkgerrors.RespondNotFound(c, "Endpoint not found")
	})

	s.engine.NoMethod(func(c *gin.Context) {
		pkgerrors.RespondProblem(c, http.StatusMethodNotAllowed, "Method Not Allowed", "Method not allowed")
	})

	s.engine.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	swagger.SetupSwagger(s.engine)

	api := s.engine.Group("/api/v1")
	{
		if s.authRouter != nil {
			s.authRouter.RegisterRoutes(api)
		}

		authRequired := api.Group("")
		authRequired.Use(middleware.AuthMiddleware(s.tokenProvider))
		{
			for _, r := range s.protectedRouters {
				if r != nil {
					r.RegisterRoutes(authRequired)
				}
			}
		}
	}
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}
