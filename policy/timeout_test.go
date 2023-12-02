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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Act
		_, err := ExecutePessimistic[string, int](
			ctx,
			func(ctx context.Context, s string) (int, error) {
				// ignore context. Does not monitor cancellation
				time.Sleep(1 * time.Hour)
				return len(s), nil
			},
			"foo",
			100*time.Millisecond,
		)

		// Assert
		if !errors.Is(err, ErrTimeoutRejected) {
			t.Fail()
		}
	})

	t.Run("should return error when passed context has already been cancelled", func(t *testing.T) {
		// Act
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Arrange
		_, err := ExecutePessimistic[string, int](
			ctx,
			func(ctx context.Context, s string) (int, error) {
				return len(s), nil
			},
			"foo",
			1*time.Hour,
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
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// Act
		_, err := ExecuteOptimistic[int, int](
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
			1*time.Hour, // > 200ms => inherit 200ms timeout
		)

		// Assert
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fail()
		}
	})

	t.Run("should return rejected by timeout error when timeout reached and parent context without timeout", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Act
		_, err := ExecuteOptimistic[int, int](
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
			200*time.Millisecond,
		)

		// Assert
		if !errors.Is(err, ErrTimeoutRejected) {
			t.Fail()
		}
	})

	t.Run("should be possible to track cancellation by timeout for graceful shutdown", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Act
		result, err := ExecuteOptimistic[int, int](
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
			50*time.Millisecond,
		)

		// Assert
		if result != 8 || err != nil {
			t.Fail()
		}
	})

	t.Run("should return error when passed context has already been canceled", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act
		_, err := ExecuteOptimistic[string, int](
			ctx,
			func(ctx context.Context, s string) (int, error) {
				return len(s), nil
			},
			"foo",
			2*time.Second,
		)

		// Assert
		if !errors.Is(err, context.Canceled) {
			t.Fail()
		}
	})
}
