package policy

import (
	"context"

	"github.com/mapogolions/resilience"
)

type FallbackPolicy[S any, T any] struct {
	fallback         func(context.Context, resilience.DelegateResult[T]) (T, error)
	resultPredicates resilience.ResultPredicates[T]
	errorPredicates  resilience.ErrorPredicates
}

func (p *FallbackPolicy[S, T]) Apply(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
	result, err := f(ctx, s)
	if err != nil {
		if p.errorPredicates.AnyMatch(err) {
			return p.fallback(ctx, resilience.DelegateResult[T]{Err: err})
		}
		return result, err
	}
	if p.resultPredicates.AnyMatch(result) {
		return p.fallback(ctx, resilience.DelegateResult[T]{Result: result})
	}
	return result, err
}

func NewFallbackPolicy[S any, T any](
	resultPredicates resilience.ResultPredicates[T],
	errorPredicates resilience.ErrorPredicates,
	fallback func(context.Context, resilience.DelegateResult[T]) (T, error),
) *FallbackPolicy[S, T] {
	return &FallbackPolicy[S, T]{
		fallback:         fallback,
		resultPredicates: resultPredicates,
		errorPredicates:  errorPredicates,
	}
}
