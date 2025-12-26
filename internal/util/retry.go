package util

import (
	"context"
	"math"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
	}
}

func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		if lastErr = fn(); lastErr == nil {
			return nil
		}

		if attempt < cfg.MaxAttempts-1 {
			delay := time.Duration(float64(cfg.BaseDelay) * math.Pow(2, float64(attempt)))
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return lastErr
}
