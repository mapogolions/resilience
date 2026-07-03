package policy

import (
	"context"
	"time"

	"github.com/mapogolions/resilience"
)

func Delay[S, T any](d time.Duration) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var zero T
		timeout, cancel := context.WithTimeout(ctx, d)
		defer cancel()

		<-timeout.Done()

		if err := ctx.Err(); err != nil {
			return zero, err
		}

		// if ctx.Err() == nil => timeout
		return f(ctx, s)
	}
}
