package policy

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDelayPolicy(t *testing.T) {
	t.Run("should be able to cancel delayed call", func(t *testing.T) {
		t.Parallel()

		// arrange
		ctx, cancel := context.WithCancel(context.Background())
		policy := Delay[string, int](24 * time.Hour)
		strlen := func(ctx context.Context, s string) (int, error) {
			return len(s), nil
		}

		// act
		time.AfterFunc(200*time.Millisecond, cancel)
		result, err := policy.Bind(strlen)(ctx, "foo")

		// assert
		if !errors.Is(err, context.Canceled) || result != 0 {
			t.Fail()
		}
	})
}
