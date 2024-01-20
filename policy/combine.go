package policy

import (
	"context"

	"github.com/mapogolions/resilience"
)

func Combine[S any, T any](g resilience.Policy[S, T], f resilience.Policy[S, T]) resilience.Policy[S, T] {
	return func(ctx context.Context, fn func(context.Context, S) (T, error), s S) (T, error) {
		return g(ctx, func(ctx context.Context, s S) (T, error) {
			return f(ctx, fn, s)
		}, s)
	}
}
