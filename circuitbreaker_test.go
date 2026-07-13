package resilience

import (
	"context"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("should return circuit broken error when failure threshold reached", func(t *testing.T) {
		// arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cb := ConsecutiveFailuresCircuitBreaker[int](1, 2*time.Second, RejectOnError)
		var f PolicyFunc[string, int] = func(ctx context.Context, s string) (int, error) {
			return 0, errSomethingWentWrong
		}

		// act
		g := f.CircuitBreaker(cb)

		result1, err1 := g(ctx, "foo")
		result2, err2 := g(ctx, "baz")

		// assert
		if result1 != 0 || err1 != errSomethingWentWrong {
			t.Fail()
		}

		if result2 != 0 || err2 != ErrCircuitBroken {
			t.Fail()
		}
	})
}
