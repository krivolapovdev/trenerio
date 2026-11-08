package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
)

type TransactionProvider interface {
	GetTransactionsBetween(ctx context.Context, startDate, endDate time.Time, ownerID uuid.UUID) ([]txmodel.Transaction, error)
}

type ScopeStrategy interface {
	Truncate(t time.Time, loc *time.Location) time.Time
	Next(t time.Time) time.Time
}

type weekStrategy struct{}

func (weekStrategy) Truncate(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	daysToMonday := weekday - 1
	d := t.AddDate(0, 0, -daysToMonday)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
}

func (weekStrategy) Next(t time.Time) time.Time {
	return t.AddDate(0, 0, 7)
}

type monthStrategy struct{}

func (monthStrategy) Truncate(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
}

func (monthStrategy) Next(t time.Time) time.Time {
	return t.AddDate(0, 1, 0)
}

type yearStrategy struct{}

func (yearStrategy) Truncate(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, loc)
}

func (yearStrategy) Next(t time.Time) time.Time {
	return t.AddDate(1, 0, 0)
}

var scopeStrategies = map[Scope]ScopeStrategy{
	Week:  weekStrategy{},
	Month: monthStrategy{},
	Year:  yearStrategy{},
}

type Service struct {
	transactionProvider TransactionProvider
}

func NewService(transactionProvider TransactionProvider) *Service {
	return &Service{
		transactionProvider: transactionProvider,
	}
}
