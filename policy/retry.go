package policy

import (
	"context"

	"github.com/mapogolions/resilience"
)

type RetryCondition[T any] func(context.Context, resilience.PolicyOutcome[T], int) bool

func RetryOnError[T any](ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
	return outcome.Err != nil
}

func NewRetryPolicy[S any, T any](retryCount int, shouldRetry RetryCondition[T]) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var result T
		var err error
		var count int
		for {
			result, err = f(ctx, s)
			count++
			if count >= retryCount {
				return result, err
			}
			outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
			if !shouldRetry(ctx, outcome, count) {
				return result, err
			}
		}
	}
}
