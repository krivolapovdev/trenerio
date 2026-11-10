package http

import "github.com/google/uuid"

type UserResponse struct {
	ID    uuid.UUID `json:"id" binding:"required" format:"uuid"`
	Email string    `json:"email" binding:"required"`
}
