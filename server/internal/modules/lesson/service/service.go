package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

type VisitProvider interface {
	GetVisitsByLessonID(ctx context.Context, lessonID, ownerID uuid.UUID) ([]visitmodel.Visit, error)
	GetVisitsByLessonIDTx(ctx context.Context, qtx *db.Queries, lessonID, ownerID uuid.UUID) ([]visitmodel.Visit, error)
	CreateVisits(ctx context.Context, lessonID uuid.UUID, studentVisits []visitmodel.StudentVisit, ownerID uuid.UUID) error
	CreateVisitsTx(ctx context.Context, qtx *db.Queries, lessonID uuid.UUID, studentVisits []visitmodel.StudentVisit, ownerID uuid.UUID) error
	UpdateVisitsByLessonID(ctx context.Context, lessonID uuid.UUID, requests []visitmodel.StudentVisit, ownerID uuid.UUID) error
	UpdateVisitsByLessonIDTx(ctx context.Context, qtx *db.Queries, lessonID uuid.UUID, requests []visitmodel.StudentVisit, ownerID uuid.UUID) ([]visitmodel.Visit, error)
	RemoveVisitsByLessonID(ctx context.Context, lessonID, ownerID uuid.UUID) error
}

type StudentProvider interface {
	GetStudentsByIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) ([]studentmodel.Student, error)
}

type Service struct {
	db              *db.Queries
	sqlDB           *sql.DB
	visitProvider   VisitProvider
	studentProvider StudentProvider
}

func NewService(
	queries *db.Queries,
	sqlDB *sql.DB,
	visitProvider VisitProvider,
	studentProvider StudentProvider,
) *Service {
	return &Service{
		db:              queries,
		sqlDB:           sqlDB,
		visitProvider:   visitProvider,
		studentProvider: studentProvider,
	}
}

func (s *Service) withTx(ctx context.Context, fn func(qtx *db.Queries) error) error {
	return db.WithTx(ctx, s.sqlDB, s.db, fn)
}
