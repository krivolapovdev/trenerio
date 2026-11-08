package domain

import "time"

// @name PaymentMetricResponse
type PaymentMetric struct {
	Date  time.Time `json:"date" binding:"required" format:"date-time"`
	Total float64   `json:"total" binding:"required"`
}
