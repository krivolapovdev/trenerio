package mapper

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/db"
	studentmodel "github.com/krivolapovdev/trenerio/internal/modules/student/domain"
	"github.com/stretchr/testify/assert"
)

func TestToRosterModel(t *testing.T) {
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

	students := []studentmodel.Student{
		{ID: uuid.New(), FullName: "Alice"},
	}

	resp := ToRosterModel(r, int64(len(students)), students)
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
	respEmpty := ToRosterModel(rEmpty, 0, nil)
	assert.Nil(t, respEmpty.Date)
	assert.Nil(t, respEmpty.Note)
	assert.Equal(t, int64(0), respEmpty.CountOfStudents)
	assert.Empty(t, respEmpty.Students)
}
