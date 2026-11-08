package mapper

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/response"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestToRosterResponse(t *testing.T) {
	rosterID := uuid.New()
	now := time.Now()
	date := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	noteText := "Test Note"

	r := db.RostersModuleRoster{
		ID:        rosterID,
		Name:      "Morning Roster",
		Date:      sql.NullTime{Time: date, Valid: true},
		Note:      sql.NullString{String: noteText, Valid: true},
		CreatedAt: now,
	}

	students := []response.StudentResponse{
		{ID: uuid.New(), FullName: "Alice"},
	}

	resp := ToRosterResponse(r, int64(len(students)), students)
	assert.Equal(t, rosterID, resp.ID)
	assert.Equal(t, "Morning Roster", resp.Name)
	assert.NotNil(t, resp.Date)
	assert.Equal(t, date, *resp.Date)
	assert.NotNil(t, resp.Note)
	assert.Equal(t, noteText, *resp.Note)
	assert.Equal(t, now, resp.CreatedAt)
	assert.Equal(t, int64(1), resp.CountOfStudents)
	assert.Len(t, resp.Students, 1)

	// Test with null fields and nil students
	rEmpty := db.RostersModuleRoster{
		ID:        rosterID,
		Name:      "Empty Roster",
		CreatedAt: now,
	}
	respEmpty := ToRosterResponse(rEmpty, 0, nil)
	assert.Nil(t, respEmpty.Date)
	assert.Nil(t, respEmpty.Note)
	assert.Empty(t, respEmpty.Students)
	assert.Equal(t, int64(0), respEmpty.CountOfStudents)
}
