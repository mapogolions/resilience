package policy

import (
	"context"
	"time"

	"github.com/mapogolions/resilience"
)

func Delay[S, T any](d time.Duration) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var zero T

		timer := time.NewTimer(d)
		defer timer.Stop()

		select {
		case <-timer.C:
			return f(ctx, s)

		case <-ctx.Done():
			return zero, ctx.Err()
		}
	}
}
