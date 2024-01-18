package policy

import (
	"context"
	"testing"
	"time"
)

func stringLengthContext(ctx context.Context, s string) (int, error) {
	return len(s), nil
}

func TestRateLimitPolicy(t *testing.T) {
	t.Run("should return true when there is free tokens", func(t *testing.T) {
		policy := NewRateLimitPolicy[string, int](1*time.Second, 1)
		v, err := policy(context.Background(), stringLengthContext, "foo")
		if err != nil || v != 3 {
			t.Fail()
		}
	})
}
