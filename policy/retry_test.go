package policy

import (
	"context"
	"sync"
	"testing"
)

func TestRetry(t *testing.T) {
	t.Run("should be able to execute policy multiple times from N threads", func(t *testing.T) {
		// Arrange
		retryCount := 3
		policy := NewRetryPolicy[string, int](retryCount, RetryOnErrorCondition)

		// Act + Assert
		wg := sync.WaitGroup{}
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				var attempts int
				defer wg.Done()
				result, err := policy(context.Background(), func(ctx context.Context, s string) (int, error) {
					attempts++
					return 0, errSomethingWentWrong
				}, "foo")

				if result != 0 || err != errSomethingWentWrong {
					t.Fail()
				}
			}()
		}
		wg.Wait()
	})

	t.Run("should break retry flow when any call succeeds", func(t *testing.T) {
		// Arrange
		var attempts int
		retryCount := 3
		policy := NewRetryPolicy[string, int](retryCount, RetryOnErrorCondition)

		// Act
		result, err := policy(context.Background(), func(ctx context.Context, s string) (int, error) {
			attempts++
			if attempts == 2 {
				return len(s), nil
			}
			return 0, errSomethingWentWrong
		}, "foo")

		// Assert
		if err != nil || result != 3 {
			t.Fail()
		}
		if attempts != 2 { // < retryCount
			t.Fail()
		}
	})

	t.Run("should retry specified amount of times when function returns an error", func(t *testing.T) {
		// Arrange
		var attempts int
		retryCount := 3
		policy := NewRetryPolicy[string, int](retryCount, RetryOnErrorCondition)

		// Act
		result, err := policy(context.Background(), func(ctx context.Context, s string) (int, error) {
			attempts++
			return 0, errSomethingWentWrong
		}, "foo")

		// Assert
		if err != errSomethingWentWrong || result != 0 {
			t.Fail()
		}
		if attempts != retryCount {
			t.Fail()
		}
	})
}
