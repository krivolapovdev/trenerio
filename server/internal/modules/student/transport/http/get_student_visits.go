package http

import (
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
)

// GetStudentVisits
// @ID           getStudentVisits
// @Tags         student-controller
// @Security     BearerAuth
// @Param        studentId  path      string  true   "Student ID"             Format(uuid)
// @Param        from       query     string  false  "Start date (RFC3339)"   Format(date-time)
// @Param        to         query     string  false  "End date (RFC3339)"     Format(date-time)
// @Success      200        {array}   VisitWithLessonResponse
// @Router       /api/v1/students/{studentId}/visits [get]
func (h *Handler) GetStudentVisits(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	studentID, ok := web.ParseUUIDParam(c, "studentId")
	if !ok {
		return
	}

	var fromPtr, toPtr *time.Time
	if fromStr := c.Query("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			pkgerrors.RespondBadRequest(c, "invalid from date format")
			return
		}
		fromPtr = &t
	}
	if toStr := c.Query("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			pkgerrors.RespondBadRequest(c, "invalid to date format")
			return
		}
		toPtr = &t
	}

	visits, err := h.service.GetStudentVisits(c.Request.Context(), studentID, jwtUser.ID, fromPtr, toPtr)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	resp := make([]VisitWithLessonResponse, len(visits))
	for i, v := range visits {
		resp[i] = VisitWithLessonResponse{
			Visit:  v.Visit,
			Lesson: v.Lesson,
		}
	}

	c.JSON(nethttp.StatusOK, resp)
}
