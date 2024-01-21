package policy

import (
	"context"
	"testing"

	"github.com/mapogolions/resilience/internal"
)

func TestRetryCount(t *testing.T) {
	t.Run("should break retry flow when any call succeeds", func(t *testing.T) {
		// Arrange
		counter := internal.NewCounter()
		retryCount := 3
		policy := NewRetryCountPolicy[string, int](retryCount, RetryOnError)

		// Act
		result, err := policy(context.Background(), func(ctx context.Context, s string) (int, error) {
			if counter.Increment() == 2 {
				return len(s), nil
			}
			return 0, errSomethingWentWrong
		}, "foo")

		// Assert
		if err != nil || result != 3 {
			t.Fail()
		}
		if int(counter.Value()) != 2 { // < retryCount
			t.Fail()
		}
	})

	t.Run("should retry specified amount of times when function returns an error", func(t *testing.T) {
		// Arrange
		counter := internal.NewCounter()
		retryCount := 3
		policy := NewRetryCountPolicy[string, int](retryCount, RetryOnError)

		// Act
		result, err := policy(context.Background(), func(ctx context.Context, s string) (int, error) {
			counter.Increment()
			return 0, errSomethingWentWrong
		}, "foo")

		// Assert
		if err != errSomethingWentWrong || result != 0 {
			t.Fail()
		}
		if int(counter.Value()) != retryCount {
			t.Fail()
		}
	})
}
