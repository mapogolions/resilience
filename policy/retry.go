package policy

import (
	"context"
	"sync/atomic"

	"github.com/mapogolions/resilience"
	"github.com/mapogolions/resilience/internal"
)

type OnRetryFunc[S any, T any] func(context.Context, resilience.PolicyOutcome[T])

func NewRetryPolicyL[S any, T any](retryCount int, onRetry OnRetryFunc[S, T]) resilience.Policy[S, T] {
	var attempts atomic.Int64
	shouldRetry := func(outcome resilience.PolicyOutcome[T]) bool {
		c := attempts.Add(1)
		return int(c) >= retryCount
	}
	var wrapOnRetry OnRetryFunc[S, T] = func(ctx context.Context, outcome resilience.PolicyOutcome[T]) {
		outcome.Dict[internal.RetryAttemptsKey] = int(attempts.Load())
		onRetry(ctx, outcome)
	}
	return NewRetryPolicyByPredicate[S, T](shouldRetry, wrapOnRetry)
}

func NewRetryPolicyByPredicate[S any, T any](
	shouldRetry func(resilience.PolicyOutcome[T]) bool,
	onRetry OnRetryFunc[S, T],
) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var result T
		var err error
		for {
			result, err = f(ctx, s)
			outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
			if !shouldRetry(outcome) {
				return result, err
			}
			onRetry(ctx, outcome)
		}
	}
}
