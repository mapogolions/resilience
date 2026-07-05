package policy

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	t.Run("should stop retrying when context is canceled", func(t *testing.T) {
		t.Parallel()

		// arrange
		retryCount := math.MaxInt
		shouldRetry := NewRetryCountOnErrorWithDelayCondition[int](retryCount, func(i int) time.Duration {
			return time.Duration((i + 1) * int(time.Second))
		})
		f := newSliceIndexer[int](nil)
		g := NewRetryPolicy[int](shouldRetry).Bind(f)
		ctx, cancel := context.WithCancel(context.Background())

		// act
		cancel()
		result, err := g(ctx, 10)

		// assert
		if !errors.Is(err, context.Canceled) || result != 0 {
			t.Fail()
		}
	})

	t.Run("should be possible to configure delay that depends on retries", func(t *testing.T) {
		// Arrange
		retryCount := 4
		shouldRetry := NewRetryCountOnErrorWithDelayCondition[int](retryCount, func(retries int) time.Duration {
			return time.Duration((retries + 1) * int(time.Millisecond))
		})
		policy := NewRetryPolicy[string](shouldRetry)

		// Act
		start := time.Now()
		policy(context.Background(), func(ctx context.Context, s string) (int, error) {
			return 0, errSomethingWentWrong
		}, "foo")
		elapsed := time.Since(start)

		if elapsed < 10*time.Millisecond {
			t.Fail()
		}
	})

	t.Run("should be able to execute policy multiple times from N threads", func(t *testing.T) {
		// Arrange
		retryCount := 3
		expectedCalls := retryCount + 1
		shouldRetry := NewRetryCountOnErrorCondition[int](retryCount)
		policy := NewRetryPolicy[string](shouldRetry)

		// Act + Assert
		wg := sync.WaitGroup{}
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var calls int
				result, err := policy(context.Background(), func(ctx context.Context, s string) (int, error) {
					calls++
					return 0, errSomethingWentWrong
				}, "foo")

				if result != 0 || err != errSomethingWentWrong || calls != expectedCalls {
					t.Fail()
				}
			}()
		}
		wg.Wait()
	})

	t.Run("should break retry flow when any call succeeds", func(t *testing.T) {
		// Arrange
		var calls int
		retryCount := 3
		shouldRetry := NewRetryCountOnErrorCondition[int](retryCount)
		policy := NewRetryPolicy[string](shouldRetry)

		// Act
		result, err := policy(context.Background(), func(ctx context.Context, s string) (int, error) {
			calls++
			if calls == 2 {
				return len(s), nil
			}
			return 0, errSomethingWentWrong
		}, "foo")

		// Assert
		if err != nil || result != 3 {
			t.Fail()
		}
		if calls != 2 {
			t.Fail()
		}
	})

	t.Run("should retry specified amount of times when function returns an error", func(t *testing.T) {
		// Arrange
		var calls int
		retryCount := 3
		expectedCalls := retryCount + 1
		shouldRetry := NewRetryCountOnErrorCondition[int](retryCount)
		policy := NewRetryPolicy[string](shouldRetry)

		// Act
		result, err := policy(context.Background(), func(ctx context.Context, s string) (int, error) {
			calls++
			return 0, errSomethingWentWrong
		}, "foo")

		// Assert
		if err != errSomethingWentWrong || result != 0 {
			t.Fail()
		}
		if calls != expectedCalls {
			t.Fail()
		}
	})
}
