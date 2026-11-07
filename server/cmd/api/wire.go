package main

import (
	"database/sql"

	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/auth"
	"github.com/krivolapovdev/trenerio/internal/modules/group"
	"github.com/krivolapovdev/trenerio/internal/modules/lesson"
	"github.com/krivolapovdev/trenerio/internal/modules/metric"
	"github.com/krivolapovdev/trenerio/internal/modules/roster"
	"github.com/krivolapovdev/trenerio/internal/modules/student"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction"
	"github.com/krivolapovdev/trenerio/internal/modules/user"
	"github.com/krivolapovdev/trenerio/internal/modules/visit"
	"github.com/krivolapovdev/trenerio/internal/server"
)

func wireServer(dbConn *sql.DB, cfg Config) *server.Server {
	queries := db.New(dbConn)
	tokenProvider := auth.NewTokenProvider(
		cfg.JWTSecret,
		cfg.AccessTokenExpiry,
		cfg.RefreshTokenExpiry,
		cfg.ReviewerKey,
	)

	userService := user.NewService(queries)
	visitService := visit.NewService(queries, dbConn)
	authService := auth.NewService(
		userService,
		tokenProvider,
		cfg.GoogleClientID,
	)

	studentService := student.NewService(queries, dbConn)

	txService := transaction.NewService(
		queries,
		dbConn,
		studentService,
	)

	lessonService := lesson.NewService(
		queries,
		dbConn,
		visitService,
		studentService,
		txService,
	)

	groupService := group.NewService(
		queries,
		dbConn,
		studentService,
		visitService,
		txService,
		lessonService,
	)

	studentService.SetDependencies(
		visitService,
		txService,
		groupService,
		lessonService,
	)

	metricService := metric.NewService(txService)
	rosterService := roster.NewService(queries, dbConn, studentService)

	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	groupHandler := group.NewHandler(groupService)
	studentHandler := student.NewHandler(studentService)
	lessonHandler := lesson.NewHandler(lessonService)
	txHandler := transaction.NewHandler(txService)
	metricHandler := metric.NewHandler(metricService)
	rosterHandler := roster.NewHandler(rosterService)

	return server.NewServer(
		tokenProvider,
		cfg.AllowedOrigins,
		auth.NewRouter(authHandler),
		user.NewRouter(userHandler),
		group.NewRouter(groupHandler),
		student.NewRouter(studentHandler),
		lesson.NewRouter(lessonHandler),
		transaction.NewRouter(txHandler),
		metric.NewRouter(metricHandler),
		roster.NewRouter(rosterHandler),
	)
}
