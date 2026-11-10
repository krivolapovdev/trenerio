package domain

import "github.com/google/uuid"

// @name User
type User struct {
	ID    uuid.UUID
	Email string
}
