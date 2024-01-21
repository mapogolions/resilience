package policy

import (
	"context"
	"testing"
)

func TestRetryCount(t *testing.T) {
	t.Run("should break retry flow when any call succeeds", func(t *testing.T) {
		// Arrange
		var attempts int
		retryCount := 3
		policy := NewRetryPolicy[string, int](retryCount, RetryOnError)

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
		policy := NewRetryPolicy[string, int](retryCount, RetryOnError)

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
