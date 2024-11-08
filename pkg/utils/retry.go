package utils

import (
	"time"
)

// RetryConfig configuration for retrying operations
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Factor       float64
}
