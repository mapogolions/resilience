package resilience

import (
	"context"
	"time"
)

func (pf PolicyFunc[S, T]) Delay(d time.Duration) PolicyFunc[S, T] {
	return Delay[S, T](d).Bind(pf)
}

func Delay[S, T any](d time.Duration) Policy[S, T] {
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
