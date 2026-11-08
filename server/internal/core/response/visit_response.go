package response

import (
	"github.com/krivolapovdev/trenerio/internal/modules/visit/model"
)

// @name VisitResponse
type VisitResponse = model.VisitResponse

// @name VisitWithLessonResponse
type VisitWithLessonResponse struct {
	Visit  VisitResponse  `json:"visit" binding:"required"`
	Lesson LessonResponse `json:"lesson" binding:"required"`
}
