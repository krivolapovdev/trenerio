package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	txhttp "github.com/krivolapovdev/trenerio/internal/modules/transaction/transport/http"
)

// GetStudentPayments
// @ID           getStudentPayments
// @Tags         student-controller
// @Security     BearerAuth
// @Param        studentId  path      string  true  "Student ID"  Format(uuid)
// @Success      200        {array}   txhttp.TransactionResponse
// @Router       /api/v1/students/{studentId}/payments [get]
func (h *Handler) GetStudentPayments(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	studentID, ok := web.ParseUUIDParam(c, "studentId")
	if !ok {
		return
	}

	payments, err := h.service.GetStudentPayments(c.Request.Context(), studentID, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	resp := make([]txhttp.TransactionResponse, len(payments))
	for i, p := range payments {
		resp[i] = txhttp.ToTransactionResponse(p)
	}

	c.JSON(nethttp.StatusOK, resp)
}
