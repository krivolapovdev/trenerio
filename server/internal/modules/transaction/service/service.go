package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
)

type StudentProvider interface {
	GetStudentsByIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) ([]studentmodel.Student, error)
	ValidateStudentExistsByID(ctx context.Context, studentID, ownerID uuid.UUID) error
}

type Service struct {
	db              *db.Queries
	sqlDB           *sql.DB
	studentProvider StudentProvider
}

func NewService(
	queries *db.Queries,
	sqlDB *sql.DB,
	studentProvider StudentProvider,
) *Service {
	return &Service{
		db:              queries,
		sqlDB:           sqlDB,
		studentProvider: studentProvider,
	}
}

func (s *Service) withTx(ctx context.Context, fn func(qtx *db.Queries) error) error {
	return db.WithTx(ctx, s.sqlDB, s.db, fn)
}
