package resilience

import (
	"context"
)

func (pf PolicyFunc[S, T]) Identity() PolicyFunc[S, T] {
	return pf
}

func NewIdentityPolicy[S any, T any]() Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		return f(ctx, s)
	}
}
