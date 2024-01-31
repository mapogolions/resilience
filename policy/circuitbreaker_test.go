package policy

import (
	"context"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("should return circuit broken error when failure threshold reached", func(t *testing.T) {
		failureThreshold := 1
		breakDuration := 2 * time.Second
		cb := NewConsecutiveFailuresCircuitBreaker[int](failureThreshold, breakDuration, RejectOnError)
		policy := NewCircuitBreakerPolicy[string, int](cb)
		f := func(ctx context.Context, s string) (int, error) {
			return 0, errSomethingWentWrong
		}

		result1, err1 := policy(context.Background(), f, "foo")
		result2, err2 := policy(context.Background(), f, "baz")

		if result1 != 0 || err1 != errSomethingWentWrong {
			t.Fail()
		}

		if result2 != 0 || err2 != ErrCircuitBroken {
			t.Fail()
		}
	})
}
