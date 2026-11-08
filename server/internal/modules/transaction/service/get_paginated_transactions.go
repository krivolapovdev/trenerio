package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
	"github.com/krivolapovdev/trenerio/internal/core/logger"
	"github.com/krivolapovdev/trenerio/internal/db"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/domain"
	"github.com/krivolapovdev/trenerio/internal/modules/transaction/mapper"
)

func (s *Service) GetPaginatedTransactions(ctx context.Context, page, size int, ownerID uuid.UUID) (domain.PageTransaction, error) {
	logger.InfoContext(ctx, "Getting paginated transactions", "page", page, "size", size, "user_id", ownerID.String())
	if size < 1 {
		logger.WarnContext(ctx, "Invalid page size for transactions", "size", size, "user_id", ownerID.String())
		return domain.PageTransaction{}, pkgerrors.NewBadRequest(fmt.Sprintf("size must be at least 1, but was %d", size))
	}

	pageIndex := page - 1
	if pageIndex < 0 {
		pageIndex = 0
	}
	offset := int32(pageIndex * size)

	totalElements, err := s.db.CountTransactionsByOwnerID(ctx, ownerID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count transactions", "error", err, "user_id", ownerID.String())
		return domain.PageTransaction{}, err
	}

	rows, err := s.db.FindTransactionsByOwnerIDPaginated(ctx, db.FindTransactionsByOwnerIDPaginatedParams{
		OwnerID: ownerID,
		Limit:   int32(size),
		Offset:  offset,
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch paginated transactions", "error", err, "offset", offset, "limit", size, "user_id", ownerID.String())
		return domain.PageTransaction{}, err
	}

	studentIDs := make([]uuid.UUID, 0, len(rows))
	seenStudentIDs := make(map[uuid.UUID]bool, len(rows))
	for _, r := range rows {
		if r.StudentID.Valid && !seenStudentIDs[r.StudentID.UUID] {
			seenStudentIDs[r.StudentID.UUID] = true
			studentIDs = append(studentIDs, r.StudentID.UUID)
		}
	}

	studentNames := make(map[uuid.UUID]string, len(studentIDs))
	if len(studentIDs) > 0 && s.studentProvider != nil {
		stRows, err := s.studentProvider.GetStudentsByIDs(ctx, studentIDs, ownerID)
		if err != nil {
			logger.WarnContext(ctx, "Failed to fetch student names for transactions", "error", err, "user_id", ownerID.String())
		} else {
			for _, sr := range stRows {
				studentNames[sr.ID] = sr.FullName
			}
		}
	}

	transactions := make([]domain.Transaction, 0, len(rows))
	for _, r := range rows {
		var details *domain.StudentPaymentDetails
		if r.StudentID.Valid && (r.PaidUntil.Valid || r.SessionsAdded.Valid) {
			var studentName *string
			if name, ok := studentNames[r.StudentID.UUID]; ok {
				studentName = &name
			}
			details = mapper.ToStudentPaymentDetails(r.StudentID.UUID, r.PaidUntil, r.SessionsAdded, studentName)
		}

		transactions = append(transactions, mapper.ToTransactionFromPaginatedRow(r, details))
	}

	totalPages := 0
	if size > 0 {
		totalPages = int((totalElements + int64(size) - 1) / int64(size))
	}
	numberOfElements := len(transactions)
	isFirst := pageIndex == 0
	isLast := pageIndex >= totalPages-1 || totalPages == 0

	logger.InfoContext(ctx, "Retrieved paginated transactions",
		"page", page,
		"size", size,
		"total_elements", totalElements,
		"returned_count", numberOfElements,
		"user_id", ownerID.String(),
	)

	return domain.PageTransaction{
		TotalPages:       totalPages,
		TotalElements:    totalElements,
		Size:             size,
		Content:          transactions,
		Number:           pageIndex,
		First:            isFirst,
		Last:             isLast,
		NumberOfElements: numberOfElements,
		Empty:            numberOfElements == 0,
	}, nil
}
