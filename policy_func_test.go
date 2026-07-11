package resilience

import (
	"context"
	"errors"
	"testing"
)

func TestPolicyFunc(t *testing.T) {
	t.Run("should not retry when fallback handles error first", func(t *testing.T) {
		t.Parallel()

		// arrange
		var calls int
		var strlen PolicyFunc[string, int] = func(ctx context.Context, s string) (int, error) {
			calls++
			return 0, errSomethingWentWrong
		}
		f := strlen.
			Fallback(func(ctx context.Context, _ int, err error) (int, error) {
				if errors.Is(err, errSomethingWentWrong) {
					return 10, nil
				}
				panic("unreachable code")
			}).
			Retry(RetryOnError[int](2))

		// act
		result, err := f(context.Background(), "foo")

		// assert
		if err != nil || result != 10 || calls != 1 {
			t.Fail()
		}

	})

	t.Run("should retry before fallback when retry wraps policy", func(t *testing.T) {
		t.Parallel()

		// arrange
		var calls int
		var strlen PolicyFunc[string, int] = func(ctx context.Context, s string) (int, error) {
			calls++
			return 0, errSomethingWentWrong
		}
		f := strlen.
			Retry(RetryOnError[int](2)).
			Fallback(func(ctx context.Context, _ int, err error) (int, error) {
				if errors.Is(err, errSomethingWentWrong) {
					return 10, nil
				}
				panic("unreachable code")
			})

		// act
		result, err := f(context.Background(), "foo")

		// assert
		if err != nil || calls != 3 || result != 10 {
			t.Fail()
		}

	})
}
