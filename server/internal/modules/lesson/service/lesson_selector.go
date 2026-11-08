package service

import (
	"github.com/google/uuid"
	lessonmodel "github.com/krivolapovdev/trenerio/internal/modules/lesson/domain"
)

// FindLatestLesson finds the most recent lesson in the map ordered by Date descending,
// breaking ties by CreatedAt descending. If the map is empty, it returns nil.
func FindLatestLesson(lessons map[uuid.UUID]lessonmodel.Lesson) *lessonmodel.Lesson {
	var latest *lessonmodel.Lesson
	for _, l := range lessons {
		if latest == nil || isLessonAfter(l, *latest) {
			cur := l
			latest = &cur
		}
	}
	return latest
}

// isLessonAfter returns true if lesson a occurred after lesson b,
// comparing by Date first and tie-breaking by CreatedAt.
func isLessonAfter(a, b lessonmodel.Lesson) bool {
	if a.Date.After(b.Date) {
		return true
	}
	return a.Date.Equal(b.Date) && a.CreatedAt.After(b.CreatedAt)
}
