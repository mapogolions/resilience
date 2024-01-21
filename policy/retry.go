package policy

import (
	"context"
	"sync/atomic"

	"github.com/mapogolions/resilience"
)

type RetryCondition[S any, T any] func(context.Context, resilience.PolicyOutcome[T], S) bool

func RetryOnError[T any](ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
	return outcome.Err != nil
}

func NewRetryCountPolicy[S any, T any](retryCount int, shouldRetry RetryCondition[int, T]) resilience.Policy[S, T] {
	condition := newRetryCountCondition[T](shouldRetry)
	return NewRetryPolicy[S, T, int](retryCount, condition)
}

func newRetryCountCondition[T any](shouldRetry RetryCondition[int, T]) RetryCondition[int, T] {
	var attempts atomic.Int64
	return func(ctx context.Context, outcome resilience.PolicyOutcome[T], retryCount int) bool {
		cur := int(attempts.Add(1))
		if cur >= retryCount {
			return false
		}
		return shouldRetry(ctx, outcome, cur)
	}
}

func NewRetryPolicy[S any, T any, R any](state R, shouldRetry RetryCondition[R, T]) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var result T
		var err error
		for {
			result, err = f(ctx, s)
			outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
			if !shouldRetry(ctx, outcome, state) {
				return result, err
			}
		}
	}
}
