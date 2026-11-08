package http

import (
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/web"
	authhttp "github.com/krivolapovdev/trenerio/internal/modules/auth/transport/http"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	txsvc "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
)

// @name CreateStudentPaymentDetailsRequest
type CreateStudentPaymentDetailsRequest struct {
	DetailsType   string     `json:"detailsType,omitempty"`
	StudentID     uuid.UUID  `json:"studentId" binding:"required" format:"uuid"`
	PaidUntil     *time.Time `json:"paidUntil,omitempty" format:"date-time"`
	SessionsAdded *int64     `json:"sessionsAdded,omitempty" binding:"omitempty,gt=0"`
}

// @name CreateTransactionRequest
type CreateTransactionRequest struct {
	Amount         float64                             `json:"amount" binding:"required,gt=0"`
	Date           time.Time                           `json:"date" binding:"required" format:"date-time"`
	Type           domain.TransactionType              `json:"type" binding:"required"`
	PaymentDetails *CreateStudentPaymentDetailsRequest `json:"paymentDetails,omitempty"`
}

// CreateTransaction
// @ID           createTransaction
// @Tags         transaction-controller
// @Security     BearerAuth
// @Param        body  body      CreateTransactionRequest  true  "Create transaction request"
// @Success      201   {object}  TransactionResponse
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

	var details *domain.StudentPaymentDetails
	if req.PaymentDetails != nil {
		details = &domain.StudentPaymentDetails{
			DetailsType:   req.PaymentDetails.DetailsType,
			StudentID:     req.PaymentDetails.StudentID,
			PaidUntil:     req.PaymentDetails.PaidUntil,
			SessionsAdded: req.PaymentDetails.SessionsAdded,
		}
	}

	txToCreate := domain.Transaction{
		Amount:         req.Amount,
		Date:           req.Date,
		Type:           req.Type,
		PaymentDetails: details,
	}

	createdTx, err := h.service.CreateTransaction(c.Request.Context(), txToCreate, jwtUser.ID)
	if err != nil {
		pkgerrors.HandleError(c, err)
		return
	}

	c.JSON(nethttp.StatusCreated, ToTransactionResponse(createdTx))
}
