package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	txmodel "github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	txservice "github.com/krivolapovdev/trenerio/internal/modules/transaction/service"
)

func (s *Service) CreateStudentPayment(ctx context.Context, studentID, ownerID uuid.UUID, amount float64, date time.Time, paidUntil *time.Time, sessionsAdded *int64) (txmodel.Transaction, error) {
	logger.InfoContext(ctx, "Creating payment for student", "student_id", studentID.String(), "amount", amount, "user_id", ownerID.String())
	if err := txservice.ValidateStudentPaymentFields(paidUntil, sessionsAdded); err != nil {
		logger.WarnContext(ctx, "Invalid student payment fields", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return txmodel.Transaction{}, err
	}

	if err := validator.ValidateDateNotFuture(date, "date"); err != nil {
		logger.WarnContext(ctx, "Student payment date cannot be in the future", "date", date.Format(time.RFC3339), "user_id", ownerID.String())
		return txmodel.Transaction{}, err
	}

	if _, err := s.GetStudentDetailsByID(ctx, studentID, ownerID); err != nil {
		logger.WarnContext(ctx, "Student not found for creating payment", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return txmodel.Transaction{}, err
	}

	var payment txmodel.Transaction
	if s.transactionProvider != nil {
		var err error
		payment, err = s.transactionProvider.CreateStudentPayment(ctx, studentID, amount, date, paidUntil, sessionsAdded, ownerID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create payment for student", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return txmodel.Transaction{}, err
		}
	}

	logger.InfoContext(ctx, "Student payment created successfully", "student_id", studentID.String(), "transaction_id", payment.ID.String(), "user_id", ownerID.String())
	return payment, nil
}
