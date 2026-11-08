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

func (s *Service) CreateStudentPayment(ctx context.Context, studentID uuid.UUID, amount float64, date time.Time, paidFrom, paidUntil *time.Time, sessionsAdded *int64, ownerID uuid.UUID) (domain.Transaction, error) {
	logger.InfoContext(ctx, "Creating student payment transaction", "student_id", studentID.String(), "amount", amount, "user_id", ownerID.String())
	if amount <= 0 {
		logger.WarnContext(ctx, "Transaction amount must be greater than 0", "amount", amount, "user_id", ownerID.String())
		return domain.Transaction{}, pkgerrors.NewBadRequest("amount must be greater than 0")
	}

	if studentID == uuid.Nil {
		logger.WarnContext(ctx, "Student ID must not be nil for student payment", "user_id", ownerID.String())
		return domain.Transaction{}, pkgerrors.NewBadRequest("studentId must not be null for student payment")
	}

	if err := ValidateStudentPaymentFieldsWithFrom(paidFrom, paidUntil, sessionsAdded); err != nil {
		logger.WarnContext(ctx, "Invalid student payment fields", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return domain.Transaction{}, err
	}

	if s.studentProvider != nil {
		if err := s.studentProvider.ValidateStudentExistsByID(ctx, studentID, ownerID); err != nil {
			logger.WarnContext(ctx, "Student not found for student payment", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return domain.Transaction{}, err
		}
	}

	if err := validator.ValidateDateNotFuture(date, "date"); err != nil {
		logger.WarnContext(ctx, "Payment date cannot be in the future", "date", date.Format(time.RFC3339), "user_id", ownerID.String())
		return domain.Transaction{}, err
	}

	var createdTx db.PaymentsModuleTransaction
	var sp db.CreateStudentPaymentRow

	err := s.withTx(ctx, func(qtx *db.Queries) error {
		amountStr := strconv.FormatFloat(amount, 'f', 2, 64)
		var err error
		createdTx, err = qtx.CreateTransaction(ctx, db.CreateTransactionParams{
			OwnerID: ownerID,
			Type:    db.PaymentsModuleTransactionTypeINCOME,
			Amount:  amountStr,
			Date:    date,
		})

		if err != nil {
			logger.ErrorContext(ctx, "Failed to create transaction record for student payment", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		effectivePaidFrom := paidFrom
		if effectivePaidFrom == nil && paidUntil != nil {
			prevPaidUntil, err := findLatestPaidUntilTx(ctx, qtx, studentID, ownerID)
			if err == nil && prevPaidUntil.Valid {
				t := prevPaidUntil.Time
				effectivePaidFrom = &t
			}
		}

		sp, err = qtx.CreateStudentPayment(ctx, db.CreateStudentPaymentParams{
			TransactionID: createdTx.ID,
			StudentID:     studentID,
			PaidFrom:      mapper.ToNullTime(effectivePaidFrom),
			PaidUntil:     mapper.ToNullTime(paidUntil),
			SessionsAdded: mapper.ToNullInt64(sessionsAdded),
		})

		if err != nil {
			logger.ErrorContext(ctx, "Failed to create student payment record", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
			return err
		}

		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create student payment transaction in database", "student_id", studentID.String(), "error", err, "user_id", ownerID.String())
		return domain.Transaction{}, err
	}

	amt, _ := strconv.ParseFloat(createdTx.Amount, 64)
	logger.InfoContext(ctx, "Student payment created successfully", "transaction_id", createdTx.ID.String(), "student_id", studentID.String(), "amount", amt, "user_id", ownerID.String())
	return mapper.ToTransaction(createdTx, mapper.ToStudentPaymentDetails(sp.StudentID, sp.PaidFrom, sp.PaidUntil, sp.SessionsAdded, nil)), nil
}
