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
		retryCondition := RetryOnError[int](math.MaxInt)
		delayProvider := func(i int) time.Duration {
			return time.Duration((i + 1) * int(time.Second))
		}
		f := newSliceIndexer[int](nil)
		g := NewRetryPolicyWithDelay[int](retryCondition, delayProvider).Bind(f)
		ctx, cancel := context.WithCancel(context.Background())

		// act
		cancel()
		result, err := g(ctx, 10) // nil[index] => error

		// assert
		if !errors.Is(err, context.Canceled) || result != 0 {
			t.Fail()
		}
	})

	t.Run("should be able to execute policy multiple times from N threads", func(t *testing.T) {
		// Arrange
		retryCount := 3
		expectedCalls := retryCount + 1
		shouldRetry := RetryOnError[int](retryCount)
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

	t.Run("should break retrying when call succeeds", func(t *testing.T) {
		// arrange
		var calls int
		retryCount := 3
		shouldRetry := RetryOnError[int](retryCount)
		policy := NewRetryPolicy[string](shouldRetry)

		// act
		result, err := policy(context.Background(), func(ctx context.Context, s string) (int, error) {
			calls++
			if calls == 2 {
				return len(s), nil
			}
			return 0, errSomethingWentWrong
		}, "foo")

		// assert
		if err != nil || result != 3 {
			t.Fail()
		}
		if calls != 2 {
			t.Fail()
		}
	})

	t.Run("should retry specified amount of times when function returns an error", func(t *testing.T) {
		// arrange
		var calls int
		retryCount := 3
		expectedCalls := retryCount + 1
		condition := RetryOnError[int](retryCount)
		policy := NewRetryPolicy[string](condition)

		g := policy.Bind(func(_ context.Context, s string) (int, error) {
			calls++
			return 0, errSomethingWentWrong
		})

		// act
		result, err := g(context.Background(), "foo")

		// assert
		if err != errSomethingWentWrong || result != 0 || calls != expectedCalls {
			t.Fail()
		}
	})
}
