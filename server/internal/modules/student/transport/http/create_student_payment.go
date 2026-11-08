package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/student/request"
	txservice "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
)

// CreateStudentPayment
// @ID           createStudentPayment
// @Tags         student-controller
// @Security     BearerAuth
// @Param        studentId  path      string                              true  "Student ID"              Format(uuid)
// @Param        body       body      request.CreateStudentPaymentRequest  true  "Student payment request"
// @Success      201        {object}  response.TransactionResponse
// @Router       /api/v1/students/{studentId}/payments [post]
func (h *Handler) CreateStudentPayment(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	studentID, ok := web.ParseUUIDParam(c, "studentId")
	if !ok {
		return
	}

	var req request.CreateStudentPaymentRequest
	if !web.BindJSON(c, &req) {
		return
	}

	if err := txservice.ValidateStudentPaymentFields(req.PaidUntil, req.SessionsAdded); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	payment, err := h.service.CreateStudentPayment(c.Request.Context(), studentID, jwtUser.ID, req)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, payment)
}
