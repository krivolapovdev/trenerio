package service

import (
	"context"
	"database/sql"

	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

type VisitResponse = domain.VisitResponse
type VisitStatus = domain.VisitStatus
type VisitType = domain.VisitType
type StudentVisit = domain.StudentVisit

type Service struct {
	db    *db.Queries
	sqlDB *sql.DB
}

func NewService(queries *db.Queries, sqlDB *sql.DB) *Service {
	return &Service{
		db:    queries,
		sqlDB: sqlDB,
	}
}

func NewVisitService(queries *db.Queries, sqlDB *sql.DB) *Service {
	return NewService(queries, sqlDB)
}

func (s *Service) withTx(ctx context.Context, fn func(qtx *db.Queries) error) error {
	return db.WithTx(ctx, s.sqlDB, s.db, fn)
}
