package domain

import "time"

// @name PaymentMetric
type PaymentMetric struct {
	Date  time.Time `json:"date" binding:"required" format:"date-time"`
	Total float64   `json:"total" binding:"required"`
}
