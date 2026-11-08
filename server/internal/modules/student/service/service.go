package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/types"
	"github.com/krivolapovdev/trenerio/internal/db"
	groupmodel "github.com/krivolapovdev/trenerio/internal/modules/group/domain"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	visitmodel "github.com/krivolapovdev/trenerio/internal/modules/visit/domain"
)

type VisitProvider interface {
	GetVisitsByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID][]visitmodel.Visit, error)
	GetVisitsByStudentID(ctx context.Context, studentID, ownerID uuid.UUID) ([]visitmodel.Visit, error)
}

type TransactionProvider interface {
	GetStudentBalancesByOwnerID(ctx context.Context, ownerID uuid.UUID) (map[uuid.UUID]txmodel.StudentBalance, error)
	GetTransactionsByStudentID(ctx context.Context, studentID, ownerID uuid.UUID) ([]txmodel.Transaction, error)
	CreateStudentPayment(ctx context.Context, studentID uuid.UUID, amount float64, date time.Time, paidFrom, paidUntil *time.Time, sessionsAdded *int64, ownerID uuid.UUID) (txmodel.Transaction, error)
}

type GroupProvider interface {
	GetGroupsByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, ownerID uuid.UUID) (map[uuid.UUID][]groupmodel.Group, error)
	SetStudentGroups(ctx context.Context, studentID, ownerID uuid.UUID, groupIDs []uuid.UUID) error
}

type LessonProvider interface {
	GetLessonsByIDs(ctx context.Context, lessonIDs []uuid.UUID, ownerID uuid.UUID) ([]lessonmodel.Lesson, error)
}

type StudentSummary struct {
	ID            uuid.UUID
	FullName      string
	Birthdate     *types.Date
	Phone         *string
	Note          *string
	CreatedAt     time.Time
	Free          bool
	Statuses      []studentmodel.StudentStatus
	StudentGroups []groupmodel.Group
	PaidUntil     *time.Time
	SessionsLeft  int64
}

type StudentVisitWithLesson struct {
	Visit  visitmodel.Visit
	Lesson lessonmodel.Lesson
}

type Service struct {
	db                  *db.Queries
	sqlDB               *sql.DB
	visitProvider       VisitProvider
	transactionProvider TransactionProvider
	groupProvider       GroupProvider
	lessonProvider      LessonProvider
}

func NewService(queries *db.Queries, sqlDB *sql.DB) *Service {
	return &Service{
		db:    queries,
		sqlDB: sqlDB,
	}
}

func (s *Service) SetDependencies(
	visitProvider VisitProvider,
	transactionProvider TransactionProvider,
	groupProvider GroupProvider,
	lessonProvider LessonProvider,
) {
	s.visitProvider = visitProvider
	s.transactionProvider = transactionProvider
	s.groupProvider = groupProvider
	s.lessonProvider = lessonProvider
}

func (s *Service) withTx(ctx context.Context, fn func(qtx *db.Queries) error) error {
	if s.sqlDB == nil {
		return fn(s.db)
	}
	return db.WithTx(ctx, s.sqlDB, s.db, fn)
}

func (s *Service) ComputeStudentStatuses(
	student studentmodel.Student,
	visits []visitmodel.Visit,
	balance *txmodel.StudentBalance,
	lessonsMap map[uuid.UUID]lessonmodel.Lesson,
	groupLessonsMap map[uuid.UUID]lessonmodel.Lesson,
	today time.Time,
) []studentmodel.StudentStatus {
	return ComputeStudentStatusesWithVisitLessons(student, visits, balance, lessonsMap, groupLessonsMap, today)
}
