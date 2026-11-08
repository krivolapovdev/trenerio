package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

type StudentProvider interface {
	GetStudentsByIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) ([]studentmodel.Student, error)
	ComputeStudentStatuses(st studentmodel.Student, visits []visitmodel.Visit, balance *txmodel.StudentBalance, lessonsMap map[uuid.UUID]lessonmodel.Lesson, groupLessonsMap map[uuid.UUID]lessonmodel.Lesson, today time.Time) []studentmodel.StudentStatus
}

type VisitProvider interface {
	GetVisitsByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID][]visitmodel.Visit, error)
	GetVisitsByLessonIDs(ctx context.Context, lessonIDs []uuid.UUID, ownerID uuid.UUID) ([]visitmodel.Visit, error)
}

type TransactionProvider interface {
	GetStudentBalancesByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID]txmodel.StudentBalance, error)
}

type LessonProvider interface {
	GetLessonsByGroupIDAndDateRange(ctx context.Context, groupID, ownerID uuid.UUID, from, to time.Time) ([]lessonmodel.Lesson, error)
	GetLessonsByIDs(ctx context.Context, lessonIDs []uuid.UUID, ownerID uuid.UUID) ([]lessonmodel.Lesson, error)
}

type Service struct {
	db                  *db.Queries
	sqlDB               *sql.DB
	studentProvider     StudentProvider
	visitProvider       VisitProvider
	transactionProvider TransactionProvider
	lessonProvider      LessonProvider
}

func NewService(
	queries *db.Queries,
	sqlDB *sql.DB,
	studentProvider StudentProvider,
	visitProvider VisitProvider,
	transactionProvider TransactionProvider,
	lessonProvider LessonProvider,
) *Service {
	return &Service{
		db:                  queries,
		sqlDB:               sqlDB,
		studentProvider:     studentProvider,
		visitProvider:       visitProvider,
		transactionProvider: transactionProvider,
		lessonProvider:      lessonProvider,
	}
}

func (s *Service) withTx(ctx context.Context, fn func(qtx *db.Queries) error) error {
	if s.sqlDB == nil {
		return fn(s.db)
	}
	return db.WithTx(ctx, s.sqlDB, s.db, fn)
}
