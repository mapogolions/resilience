package policy

import (
	"context"
	"errors"
	"fmt"

	"github.com/mapogolions/resilience"
)

type Fallback[T any] func(context.Context, resilience.PolicyOutcome[T]) (T, error)

func IdentityFallback[T any](ctx context.Context, outcome resilience.PolicyOutcome[T]) (T, error) {
	return outcome.Result, outcome.Err
}

func NewFallbackPolicy[S any, T any](fallback Fallback[T]) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		result, err := f(ctx, s)
		return fallback(ctx, resilience.PolicyOutcome[T]{Result: result, Err: err})
	}
}

func NewPanicFallbackPolicy[S any, T any](fallback Fallback[T]) resilience.Policy[S, T] {
	policy := NewFallbackPolicy[S, T](fallback)
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		return policy(ctx, func(ctx context.Context, s S) (T, error) {
			var crucialErr error
			var result T
			var err error
			tryCatch(func() { result, err = f(ctx, s) }, &crucialErr)
			if crucialErr != nil {
				return result, crucialErr
			}
			return result, err
		}, s)
	}
}

func tryCatch(f func(), err *error) {
	defer func() {
		if info := recover(); info != nil {
			if errorMessage, ok := info.(string); ok {
				*err = errors.New(errorMessage)
				return
			}
			if ex, ok := info.(error); ok {
				*err = ex
				return
			}
			*err = fmt.Errorf("%v", info)
		}
	}()
	f()
}
