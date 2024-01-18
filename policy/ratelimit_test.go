package policy

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRateLimitPolicy(t *testing.T) {
	t.Run("should reject execution and return error when there is no free token", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		policy := NewRateLimitPolicy[string, int](1*time.Hour, 1)
		var calls []string
		f := func(ctx context.Context, s string) (int, error) {
			calls = append(calls, s)
			return len(s), nil
		}

		// Act
		policy(ctx, f, "foo")
		_, err := policy(ctx, f, "bar")

		// Assert
		if len(calls) != 1 || calls[0] != "foo" || !errors.Is(err, ErrRateLimitRejected) {
			t.Fail()
		}
	})

	t.Run("should permit execution when there are free tokens", func(t *testing.T) {
		// Arrange
		policy := NewRateLimitPolicy[string, int](1*time.Second, 1)
		f := func(ctx context.Context, s string) (int, error) {
			return len(s), nil
		}

		// Act
		v, err := policy(context.Background(), f, "foo")

		// Assert
		if err != nil || v != 3 {
			t.Fail()
		}
	})
}
