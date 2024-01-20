package policy

import (
	"context"
	"errors"
	"fmt"

	"github.com/mapogolions/resilience"
)

type FallbackFunc[T any] func(context.Context, resilience.PolicyOutcome[T]) (T, error)

func IdentityFallback[T any](ctx context.Context, outcome resilience.PolicyOutcome[T]) (T, error) {
	return outcome.Result, outcome.Err
}

func NewFallbackPolicy[S any, T any](fallback FallbackFunc[T]) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		result, err := f(ctx, s)
		return fallback(ctx, resilience.PolicyOutcome[T]{Result: result, Err: err})
	}
}

func NewPanicFallbackPolicy[S any, T any](fallback FallbackFunc[T]) resilience.Policy[S, T] {
	policy := NewFallbackPolicy[S, T](fallback)
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		return policy(ctx, func(ctx context.Context, s S) (T, error) {
			var result T
			var crucialErr, err error
			tryCatch(func() { result, err = f(ctx, s) }, &crucialErr)
			if crucialErr != nil {
				return result, crucialErr
			}
			return result, err
		}, s)
	}
}

func tryCatch(f func(), crucialErr *error) {
	defer func() {
		if info := recover(); info != nil {
			if errorMessage, ok := info.(string); ok {
				*crucialErr = errors.New(errorMessage)
				return
			}
			if err, ok := info.(error); ok {
				*crucialErr = err
				return
			}
			*crucialErr = fmt.Errorf("%v", info)
		}
	}()
	f()
}
