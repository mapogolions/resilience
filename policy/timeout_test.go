package policy

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPessimisticTimeout(t *testing.T) {
	t.Run("should not block the flow for an indefinite period of time when user ignore context", func(t *testing.T) {
		// It is the main difference between `Pessimistic` and `Optimisitic` scenarios. Don't rely on callers
		// Arrange
		policy := NewTimeoutPolicy[string, int](100*time.Millisecond, PessimisticTimeoutPolicy)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Act
		_, err := policy.Apply(
			ctx,
			func(ctx context.Context, s string) (int, error) {
				// ignore context. Does not monitor cancellation
				time.Sleep(1 * time.Hour)
				return len(s), nil
			},
			"foo",
		)

		// Assert
		if !errors.Is(err, ErrTimeoutRejected) {
			t.Fail()
		}
	})

	t.Run("should return error when passed context has already been cancelled", func(t *testing.T) {
		// Act
		policy := NewTimeoutPolicy[string, int](1*time.Hour, PessimisticTimeoutPolicy)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Arrange
		_, err := policy.Apply(
			ctx,
			func(ctx context.Context, s string) (int, error) {
				return len(s), nil
			},
			"foo",
		)

		// Assert
		if !errors.Is(err, context.Canceled) {
			t.Fail()
		}
	})
}

func TestOptimisticTimeout(t *testing.T) {
	t.Run("should return deadline exceeded error when context inherits deadline from parent context", func(t *testing.T) {
		// Arrange
		policy := NewTimeoutPolicy[int, int](1*time.Hour /* > 200ms => inherit 200ms timeout */, OptimisticTimeoutPolicy)
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// Act
		_, err := policy.Apply(
			ctx,
			func(ctx context.Context, n int) (int, error) {
				for {
					select {
					case <-ctx.Done():
						return 0, ctx.Err()
					case <-time.After(100 * time.Millisecond):
					}
				}
			},
			10_000,
		)

		// Assert
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fail()
		}
	})

	t.Run("should return rejected by timeout error when timeout reached and parent context without timeout", func(t *testing.T) {
		// Arrange
		policy := NewTimeoutPolicy[int, int](200*time.Millisecond, OptimisticTimeoutPolicy)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Act
		_, err := policy.Apply(
			ctx,
			func(ctx context.Context, n int) (int, error) {
				for {
					select {
					case <-ctx.Done():
						return 0, ctx.Err()
					case <-time.After(100 * time.Millisecond):
					}
				}
			},
			10_000,
		)

		// Assert
		if !errors.Is(err, ErrTimeoutRejected) {
			t.Fail()
		}
	})

	t.Run("should be possible to track cancellation by timeout for graceful shutdown", func(t *testing.T) {
		// Arrange
		policy := NewTimeoutPolicy[int, int](50*time.Millisecond, OptimisticTimeoutPolicy)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Act
		result, err := policy.Apply(
			ctx,
			func(ctx context.Context, n int) (int, error) {
				for {
					select {
					case <-ctx.Done():
						return 8, nil
					case <-time.After(100 * time.Millisecond):
					}
				}
			},
			10_000,
		)

		// Assert
		if result != 8 || err != nil {
			t.Fail()
		}
	})

	t.Run("should return error when passed context has already been canceled", func(t *testing.T) {
		// Arrange
		policy := NewTimeoutPolicy[string, int](2*time.Second, OptimisticTimeoutPolicy)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act
		_, err := policy.Apply(
			ctx,
			func(ctx context.Context, s string) (int, error) {
				return len(s), nil
			},
			"foo",
		)

		// Assert
		if !errors.Is(err, context.Canceled) {
			t.Fail()
		}
	})
}
