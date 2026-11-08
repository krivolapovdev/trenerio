package service

import (
	"github.com/krivolapovdev/trenerio/internal/db"
)

type Service struct {
	db *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{
		db: queries,
	}
}
