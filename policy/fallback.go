package policy

import (
	"context"

	"github.com/mapogolions/resilience"
)

type FallbackPolicy[S any, T any] func(context.Context, func(context.Context, S) (T, error), S) (T, error)
type Fallback[T any] func(context.Context, resilience.DelegateResult[T]) (T, error)

func NewFallbackPolicy[S any, T any](
	fallback Fallback[T],
	resultPredicates resilience.ResultPredicates[T],
	errorPredicates resilience.ErrorPredicates,
) FallbackPolicy[S, T] {
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
