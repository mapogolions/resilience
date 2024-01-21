package policy

import (
	"context"
	"sync/atomic"

	"github.com/mapogolions/resilience"
)

type RetryPred[T any, R any] func(resilience.PolicyOutcome[T]) (bool, R)
type OnRetryFunc[T any, R any] func(context.Context, resilience.PolicyOutcome[T], R)

func OnRetryIdentity[S any, T any, R any](context.Context, resilience.PolicyOutcome[T], R) {}
func OnRetryCountIdentity[S any, T any](context.Context, resilience.PolicyOutcome[T], int) {}

func NewRetryCountPolicy[S any, T any](retryCount int, onRetry OnRetryFunc[T, int]) resilience.Policy[S, T] {
	var attempts atomic.Int64
	shouldRetry := func(outcome resilience.PolicyOutcome[T]) (bool, int) {
		cur := int(attempts.Add(1))
		return cur >= retryCount, cur
	}
	return NewRetryPolicy[S, T, int](shouldRetry, onRetry)
}

func NewRetryPolicy[S any, T any, R any](
	shouldRetry RetryPred[T, R],
	onRetry OnRetryFunc[T, R],
) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var result T
		var err error
		for {
			result, err = f(ctx, s)
			outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
			if ok, r := shouldRetry(outcome); ok {
				onRetry(ctx, outcome, r)
				continue
			}
			return result, err
		}
	}
}
