package policy

import (
	"context"
	"time"

	"github.com/mapogolions/resilience"
)

func Delay[S, T any](d time.Duration) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		time.Sleep(d)
		return f(ctx, s)
	}
}
