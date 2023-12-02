package policy

import (
	"context"

	"github.com/mapogolions/resilience"
)

func NewFallback[S any, T any](
	policyBuilder resilience.PolicyBuilder[T],
	fallback func(context.Context, resilience.DelegateResult[T]) (T, error),
) func(context.Context, func(context.Context, S) (T error), S) (T, error) {
	return func(ctx context.Context, f func(context.Context, S) (T error), state S) (T, error) {
		var defaultValue T
		return defaultValue, nil
	}
}
