package validator

import (
	"fmt"
	"strings"
	"time"

	pkgerrors "github.com/krivolapovdev/trenerio/internal/core/errors"
)

// ValidateRequiredName trims the input name and ensures it is not blank and does not exceed 255 characters.
func ValidateRequiredName(name, fieldName string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", pkgerrors.NewBadRequest(fmt.Sprintf("%s must not be blank", fieldName))
	}
	if len(trimmed) > 255 {
		return "", pkgerrors.NewBadRequest(fmt.Sprintf("%s must not exceed 255 characters", fieldName))
	}
	return trimmed, nil
}

// ValidateNameUpdate checks if fieldName exists in updates. If present, it validates that it's a non-blank string
// within 255 characters and updates the map with the trimmed string. Returns error if invalid.
func ValidateNameUpdate(updates map[string]any, fieldName string) error {
	raw, ok := updates[fieldName]
	if !ok || raw == nil {
		return nil
	}
	nameStr, ok := raw.(string)
	if !ok {
		return pkgerrors.NewBadRequest(fmt.Sprintf("%s must not be blank", fieldName))
	}
	trimmed, err := ValidateRequiredName(nameStr, fieldName)
	if err != nil {
		return err
	}
	updates[fieldName] = trimmed
	return nil
}

// SanitizeTrimmedStringUpdate trims the string value in updates if the field is present and non-nil.
func SanitizeTrimmedStringUpdate(updates map[string]any, fieldName string) {
	if raw, ok := updates[fieldName]; ok && raw != nil {
		if str, ok := raw.(string); ok {
			updates[fieldName] = strings.TrimSpace(str)
		}
	}
}

// TrimStringPtr returns a pointer to the trimmed string, or nil if the input pointer is nil.
func TrimStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	return &trimmed
}

// ValidateDateNotFuture ensures the date is in the present or past in UTC (not after time.Now().UTC()).
func ValidateDateNotFuture(d time.Time, fieldName string) error {
	if !d.IsZero() && d.UTC().After(time.Now().UTC()) {
		return pkgerrors.NewBadRequest(fmt.Sprintf("%s cannot be in the future", fieldName))
	}
	return nil
}
