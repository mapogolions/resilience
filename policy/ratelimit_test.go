package policy

import (
	"context"
	"testing"
	"time"
)

func TestRateLimitPolicy(t *testing.T) {
	t.Run("should permit execution when there are free tokens", func(t *testing.T) {
		policy := NewRateLimitPolicy[string, int](1*time.Second, 1)
		v, err := policy(context.Background(), stringLengthContext, "foo")

		if err != nil || v != 3 {
			t.Fail()
		}
	})
}

func stringLengthContext(ctx context.Context, s string) (int, error) {
	return len(s), nil
}
