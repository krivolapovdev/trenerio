package domain

import "github.com/google/uuid"

type JwtUser struct {
	ID    uuid.UUID
	Email string
}
