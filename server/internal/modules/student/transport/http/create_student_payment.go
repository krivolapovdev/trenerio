package http

import (
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	txservice "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
	txhttp "github.com/krivolapovdev/trenerio/internal/modules/transaction/transport/http"
)

// @name CreateStudentPaymentRequest
type CreateStudentPaymentRequest struct {
	Amount        float64    `json:"amount" binding:"required,gt=0"`
	Date          time.Time  `json:"date" binding:"required" format:"date-time"`
	PaidUntil     *time.Time `json:"paidUntil,omitempty" format:"date-time"`
	SessionsAdded *int64     `json:"sessionsAdded,omitempty" binding:"omitempty,gt=0"`
}

// CreateStudentPayment
// @ID           createStudentPayment
// @Tags         student-controller
// @Security     BearerAuth
// @Param        studentId  path      string                       true  "Student ID"              Format(uuid)
// @Param        body       body      CreateStudentPaymentRequest  true  "Student payment request"
// @Success      201        {object}  TransactionResponse
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

	var req CreateStudentPaymentRequest
	if !web.BindJSON(c, &req) {
		return
	}

	if err := txservice.ValidateStudentPaymentFields(req.PaidUntil, req.SessionsAdded); err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	payment, err := h.service.CreateStudentPayment(c.Request.Context(), studentID, jwtUser.ID, req.Amount, req.Date, req.PaidUntil, req.SessionsAdded)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, txhttp.ToTransactionResponse(payment))
}
