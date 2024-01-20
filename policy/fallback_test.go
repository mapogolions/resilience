package policy

import (
	"context"
	"errors"
	"testing"

	"github.com/mapogolions/resilience"
)

var errSomethingWentWrong = errors.New("something went wrong")

func TestPanicFallback(t *testing.T) {
	t.Run("should recovery from panic", func(t *testing.T) {
		// Arrange
		f := func(ctx context.Context, s string) (int, error) {
			panic(errSomethingWentWrong)
		}
		policy := NewPanicFallbackPolicy[string, int](IdentityFallback)

		// Act
		result, err := policy(context.Background(), f, "foo")

		// Assert
		if result != 0 || err != errSomethingWentWrong {
			t.Fail()
		}
	})
}

func TestFallback(t *testing.T) {
	t.Run("fallback should be able to ignore original error and return fallback value", func(t *testing.T) {
		// Arrange
		fallbackResult := -1
		var fallback Fallback[int] = func(ctx context.Context, outcome resilience.PolicyOutcome[int]) (int, error) {
			if outcome.Result != 0 || outcome.Err != errSomethingWentWrong {
				t.Fail()
			}
			return fallbackResult, nil
		}
		f := func(ctx context.Context, s string) (int, error) {
			return 0, errSomethingWentWrong
		}
		policy := NewFallbackPolicy[string, int](fallback)

		// Act
		result, err := policy(context.Background(), f, "foo")

		// Assert
		if result != fallbackResult || err != nil {
			t.Fail()
		}
	})

	t.Run("should return original result when IdentityFallback used", func(t *testing.T) {
		// Arrange
		f := func(ctx context.Context, s string) (int, error) {
			return len(s), nil
		}
		policy := NewFallbackPolicy[string, int](IdentityFallback)

		// Act
		result, err := policy(context.Background(), f, "foo")

		// Assert
		if result != 3 || err != nil {
			t.Fail()
		}
	})

	t.Run("should return original error when IdentityFallback used", func(t *testing.T) {
		// Arrange
		f := func(ctx context.Context, s string) (int, error) {
			return 0, errSomethingWentWrong
		}
		policy := NewFallbackPolicy[string, int](IdentityFallback)

		// Act
		result, err := policy(context.Background(), f, "foo")

		// Assert
		if result != 0 || err != errSomethingWentWrong {
			t.Fail()
		}
	})
}
