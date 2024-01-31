package policy

import (
	"context"
	"testing"
)

func TestComposition(t *testing.T) {
	t.Run("Should build a pipeline from multiple policies, with the first one being the outermost", func(t *testing.T) {
		// Arrange
		policy := Pipeline[string, int](minus, addOne, power2)

		// Act
		result, err := policy(
			context.Background(),
			func(ctx context.Context, s string) (int, error) { return len(s), nil },
			"foo",
		)

		// Assert
		if err != nil || result != -10 {
			t.Fail()
		}
	})

	t.Run("should build pipeline that consists of identity policy only", func(t *testing.T) {
		// Arrange
		policy := Pipeline[string, int]()

		// Act
		result, err := policy(
			context.Background(),
			func(ctx context.Context, s string) (int, error) { return len(s), nil },
			"foo",
		)

		// Assert
		if err != nil || result != 3 {
			t.Fail()
		}
	})

	t.Run("should combine two policies", func(t *testing.T) {
		// Arrange
		policy := Compose(addOne, power2)

		// Act
		result, err := policy(
			context.Background(),
			func(ctx context.Context, s string) (int, error) { return len(s), nil },
			"foo",
		)

		// Assert
		if err != nil || result != 10 {
			t.Fail()
		}
	})
}

func addOne(ctx context.Context, f func(context.Context, string) (int, error), s string) (int, error) {
	result, err := f(ctx, s)
	return result + 1, err
}

func power2(ctx context.Context, f func(context.Context, string) (int, error), s string) (int, error) {
	result, err := f(ctx, s)
	return result * result, err
}

func minus(ctx context.Context, f func(context.Context, string) (int, error), s string) (int, error) {
	result, err := f(ctx, s)
	return -result, err
}
