package policy

import (
	"context"

	"github.com/mapogolions/resilience"
)

func NewIdentityPolicy[S any, T any]() resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		return f(ctx, s)
	}
}
