package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	txsvc "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
)

// CreateTransaction
// @ID           createTransaction
// @Tags         transaction-controller
// @Security     BearerAuth
// @Param        body  body      request.CreateTransactionRequest  true  "Create transaction request"
// @Success      201   {object}  response.TransactionResponse
// @Router       /api/v1/transactions [post]
func (h *Handler) CreateTransaction(c *gin.Context) {
	jwtUser, ok := authhttp.GetUserFromContext(c)
	if !ok {
		return
	}

	var req CreateTransactionRequest
	if !web.BindJSON(c, &req) {
		return
	}

	if req.PaymentDetails != nil {
		if err := txsvc.ValidateStudentPaymentFields(req.PaymentDetails.PaidUntil, req.PaymentDetails.SessionsAdded); err != nil {
			pkgerrors.HandleError(c, err)
			return
		}
	}

	createdTx, err := h.service.CreateTransaction(c.Request.Context(), req, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, createdTx)
}
