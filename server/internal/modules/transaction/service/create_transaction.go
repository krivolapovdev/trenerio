package service

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/core/validator"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) CreateTransaction(ctx context.Context, txToCreate domain.Transaction, ownerID uuid.UUID) (domain.Transaction, error) {
	logger.InfoContext(ctx, "Creating transaction", "amount", txToCreate.Amount, "type", string(txToCreate.Type), "user_id", ownerID.String())
	if txToCreate.Amount <= 0 {
		logger.WarnContext(ctx, "Transaction amount must be greater than 0", "amount", txToCreate.Amount, "user_id", ownerID.String())
		return domain.Transaction{}, pkgerrors.NewBadRequest("amount must be greater than 0")
	}

	if err := validator.ValidateDateNotFuture(txToCreate.Date, "date"); err != nil {
		logger.WarnContext(ctx, "Transaction date cannot be in the future", "date", txToCreate.Date.Format(time.RFC3339), "user_id", ownerID.String())
		return domain.Transaction{}, err
	}

	if txToCreate.PaymentDetails != nil {
		if txToCreate.PaymentDetails.StudentID == uuid.Nil {
			logger.WarnContext(ctx, "Student ID must not be nil for student payment", "user_id", ownerID.String())
			return domain.Transaction{}, pkgerrors.NewBadRequest("studentId must not be null for student payment")
		}

		if err := ValidateStudentPaymentFields(txToCreate.PaymentDetails.PaidUntil, txToCreate.PaymentDetails.SessionsAdded); err != nil {
			logger.WarnContext(ctx, "Invalid student payment details", "student_id", txToCreate.PaymentDetails.StudentID.String(), "error", err, "user_id", ownerID.String())
			return domain.Transaction{}, err
		}

		if s.studentProvider != nil {
			if err := s.studentProvider.ValidateStudentExistsByID(ctx, txToCreate.PaymentDetails.StudentID, ownerID); err != nil {
				logger.WarnContext(ctx, "Student not found for transaction payment details", "student_id", txToCreate.PaymentDetails.StudentID.String(), "error", err, "user_id", ownerID.String())
				return domain.Transaction{}, err
			}
		}
	}

	var createdTx db.PaymentsModuleTransaction
	var details *domain.StudentPaymentDetails

	err := s.withTx(ctx, func(qtx *db.Queries) error {
		amountStr := strconv.FormatFloat(txToCreate.Amount, 'f', 2, 64)
		var err error
		createdTx, err = qtx.CreateTransaction(ctx, db.CreateTransactionParams{
			OwnerID: ownerID,
			Type:    db.PaymentsModuleTransactionType(txToCreate.Type),
			Amount:  amountStr,
			Date:    txToCreate.Date,
		})

		if err != nil {
			logger.ErrorContext(ctx, "Failed to create transaction record", "error", err, "user_id", ownerID.String())
			return err
		}

		if txToCreate.PaymentDetails != nil {
			sp, err := qtx.CreateStudentPayment(ctx, db.CreateStudentPaymentParams{
				TransactionID: createdTx.ID,
				StudentID:     txToCreate.PaymentDetails.StudentID,
				PaidUntil:     mapper.ToNullTime(txToCreate.PaymentDetails.PaidUntil),
				SessionsAdded: mapper.ToNullInt64(txToCreate.PaymentDetails.SessionsAdded),
			})
			if err != nil {
				logger.ErrorContext(ctx, "Failed to create student payment details record", "transaction_id", createdTx.ID.String(), "student_id", txToCreate.PaymentDetails.StudentID.String(), "error", err, "user_id", ownerID.String())
				return err
			}

			if _, err := qtx.ApplyPaymentToStudentBalance(ctx, db.ApplyPaymentToStudentBalanceParams{
				OwnerID:       ownerID,
				StudentID:     sp.StudentID,
				PaidUntil:     sp.PaidUntil,
				SessionsAdded: sp.SessionsAdded,
			}); err != nil {
				logger.ErrorContext(ctx, "Failed to update student balance on transaction creation", "transaction_id", createdTx.ID.String(), "student_id", sp.StudentID.String(), "error", err, "user_id", ownerID.String())
				return err
			}

			details = mapper.ToStudentPaymentDetails(sp.StudentID, sp.PaidUntil, sp.SessionsAdded, nil)
		}
		return nil
	})

	if err != nil {
		logger.ErrorContext(ctx, "Failed to create transaction in database", "error", err, "user_id", ownerID.String())
		return domain.Transaction{}, err
	}

	amt, _ := strconv.ParseFloat(createdTx.Amount, 64)
	logger.InfoContext(ctx, "Transaction created successfully", "transaction_id", createdTx.ID.String(), "amount", amt, "type", string(createdTx.Type), "user_id", ownerID.String())
	return mapper.ToTransaction(createdTx, details), nil
}
