package main

import (
	"database/sql"

	"github.com/krivolapovdev/trenerio/internal/db"
	authsvc "github.com/krivolapovdev/trenerio/internal/modules/auth/service"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	groupsvc "github.com/krivolapovdev/trenerio/internal/modules/group/service"
	grouphttp "github.com/krivolapovdev/trenerio/internal/modules/group/transport/http"
	lessonsvc "github.com/krivolapovdev/trenerio/internal/modules/lesson/service"
	lessonhttp "github.com/krivolapovdev/trenerio/internal/modules/lesson/transport/http"
	metricsvc "github.com/krivolapovdev/trenerio/internal/modules/metric/service"
	metrichttp "github.com/krivolapovdev/trenerio/internal/modules/metric/transport/http"
	rostersvc "github.com/krivolapovdev/trenerio/internal/modules/roster/service"
	rosterhttp "github.com/krivolapovdev/trenerio/internal/modules/roster/transport/http"
	studentsvc "github.com/krivolapovdev/trenerio/internal/modules/student/service"
	studenthttp "github.com/krivolapovdev/trenerio/internal/modules/student/transport/http"
	txsvc "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
	txhttp "github.com/krivolapovdev/trenerio/internal/modules/transaction/transport/http"
	usersvc "github.com/krivolapovdev/trenerio/internal/modules/user/service"
	userhttp "github.com/krivolapovdev/trenerio/internal/modules/user/transport/http"
	visitsvc "github.com/krivolapovdev/trenerio/internal/modules/visit/service"
	"github.com/krivolapovdev/trenerio/internal/server"
)

func wireServer(dbConn *sql.DB, cfg Config) *server.Server {
	queries := db.New(dbConn)
	tokenProvider := authsvc.NewTokenProvider(
		cfg.JWTSecret,
		cfg.AccessTokenExpiry,
		cfg.RefreshTokenExpiry,
		cfg.ReviewerKey,
	)

	userService := usersvc.NewService(queries)
	visitService := visitsvc.NewService(queries, dbConn)
	authService := authsvc.NewService(
		userService,
		tokenProvider,
		cfg.GoogleClientID,
	)

	studentService := studentsvc.NewService(queries, dbConn)

	txService := txsvc.NewService(
		queries,
		dbConn,
		studentService,
	)

	lessonService := lessonsvc.NewService(
		queries,
		dbConn,
		visitService,
		studentService,
	)

	groupService := groupsvc.NewService(
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

	metricService := metricsvc.NewService(txService)
	rosterService := rostersvc.NewService(queries, dbConn, studentService)

	authHandler := authhttp.NewHandler(authService)
	userHandler := userhttp.NewHandler(userService)
	groupHandler := grouphttp.NewHandler(groupService)
	studentHandler := studenthttp.NewHandler(studentService)
	lessonHandler := lessonhttp.NewHandler(lessonService)
	txHandler := txhttp.NewHandler(txService)
	metricHandler := metrichttp.NewHandler(metricService)
	rosterHandler := rosterhttp.NewHandler(rosterService)

	return server.NewServer(
		tokenProvider,
		cfg.AllowedOrigins,
		authHandler,
		userHandler,
		groupHandler,
		studentHandler,
		lessonHandler,
		txHandler,
		metricHandler,
		rosterHandler,
	)
}
