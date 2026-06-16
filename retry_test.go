package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetrySucceedsFirstAttempt(t *testing.T) {
	ctx := context.Background()
	var callCount atomic.Int32

	err := Retry(ctx, 3, 10*time.Millisecond, func() error {
		callCount.Add(1)
		return nil
	})

	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}
	if callCount.Load() != 1 {
		t.Errorf("expected 1 call, got %d", callCount.Load())
	}
}

func TestRetrySucceedsAfterFailures(t *testing.T) {
	ctx := context.Background()
	var callCount atomic.Int32

	err := Retry(ctx, 3, 10*time.Millisecond, func() error {
		n := callCount.Add(1)
		if n < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}
	if callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", callCount.Load())
	}
}

func TestRetryExhaustsAllAttempts(t *testing.T) {
	ctx := context.Background()
	var callCount atomic.Int32
	expectedErr := errors.New("persistent error")

	err := Retry(ctx, 3, 10*time.Millisecond, func() error {
		callCount.Add(1)
		return expectedErr
	})

	if err == nil {
		t.Error("expected error after exhausting retries")
	}
	if callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", callCount.Load())
	}
}

func TestRetryWithContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var callCount atomic.Int32

	err := Retry(ctx, 5, 100*time.Millisecond, func() error {
		callCount.Add(1)
		return errors.New("always fails")
	})

	if err == nil {
		t.Error("expected error due to context timeout")
	}
	if callCount.Load() != 1 {
		t.Errorf("expected 1 call before context cancel, got %d", callCount.Load())
	}
}

func TestRetryZeroCountDoesNothing(t *testing.T) {
	ctx := context.Background()
	var callCount atomic.Int32

	err := Retry(ctx, 0, 0, func() error {
		callCount.Add(1)
		return errors.New("error")
	})

	if err != nil {
		t.Errorf("expected nil error for zero retries, got: %v", err)
	}
	if callCount.Load() != 0 {
		t.Errorf("expected 0 calls, got %d", callCount.Load())
	}
}
