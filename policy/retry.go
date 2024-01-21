package policy

import (
	"context"
	"sync/atomic"

	"github.com/mapogolions/resilience"
)

type RetryPred[T any] func(context.Context, resilience.PolicyOutcome[T]) bool
type OnRetryFunc[T any, R any] func(context.Context, resilience.PolicyOutcome[T], R)

func OnRetryCountIdentity[T any](context.Context, resilience.PolicyOutcome[T], int) {}

func NewRetryCountPolicy[S any, T any](retryCount int, onRetry OnRetryFunc[T, int]) resilience.Policy[S, T] {
	var attempts atomic.Int64
	shouldRetry := func(ctx context.Context, outcome resilience.PolicyOutcome[T]) bool {
		cur := int(attempts.Add(1))
		if cur >= retryCount {
			return false
		}
		onRetry(ctx, outcome, cur)
		return true
	}
	return NewRetryPolicy[S, T, int](shouldRetry)
}

func NewRetryPolicy[S any, T any, R any](shouldRetry RetryPred[T]) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var result T
		var err error
		for {
			result, err = f(ctx, s)
			outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
			if !shouldRetry(ctx, outcome) {
				return result, err
			}
		}
	}
}
