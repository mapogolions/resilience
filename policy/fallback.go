package policy

import (
	"context"

	"github.com/mapogolions/resilience"
)

func NewFallback[S any, T any](
	resultPredicates resilience.ResultPredicates[T],
	errorPredicates resilience.ErrorPredicates,
	fallback func(context.Context, resilience.DelegateResult[T]) (T, error),
) func(context.Context, func(context.Context, S) (T, error), S) (T, error) {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		result, err := f(ctx, s)
		if err != nil {
			if errorPredicates.AnyMatch(err) {
				return fallback(ctx, resilience.DelegateResult[T]{Err: err})
			}
			return result, err
		}
		if resultPredicates.AnyMatch(result) {
			return fallback(ctx, resilience.DelegateResult[T]{Result: result})
		}
		return result, err
	}
}
