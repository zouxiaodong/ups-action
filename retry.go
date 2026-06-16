package main

import (
	"context"
	"fmt"
	"time"
)

// Retry calls fn up to maxAttempts times, waiting delay between each failed attempt.
// If maxAttempts is 0, it does nothing and returns nil.
// If the context is canceled or times out, Retry returns immediately.
func Retry(ctx context.Context, maxAttempts int, delay time.Duration, fn func() error) error {
	if maxAttempts <= 0 {
		return nil
	}

	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		// Check context before each attempt
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("context done after %d attempt(s): %w (last error: %v)", i, ctx.Err(), lastErr)
			}
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't sleep after the last attempt
		if i < maxAttempts-1 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context done during retry delay: %w (last error: %v)", ctx.Err(), lastErr)
			case <-time.After(delay):
			}
		}
	}

	return fmt.Errorf("all %d attempt(s) failed: %w", maxAttempts, lastErr)
}
