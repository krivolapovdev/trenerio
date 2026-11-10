package domain

import "github.com/google/uuid"

// @name JwtUser
type JwtUser struct {
	ID    uuid.UUID
	Email string
}
