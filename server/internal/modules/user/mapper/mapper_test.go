package mapper

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestToUserModel(t *testing.T) {
	userID := uuid.New()
	u := db.UsersModuleOauth2User{
		ID:         userID,
		Provider:   db.UsersModuleOauth2ProviderGOOGLE,
		ProviderID: "google-12345",
		Email:      "TEST@EXAMPLE.COM",
		CreatedAt:  time.Now(),
	}

	res := ToUserModel(u)
	assert.Equal(t, userID, res.ID)
	assert.Equal(t, "TEST@EXAMPLE.COM", res.Email)
}
