package policy

import (
	"context"
	"time"

	"github.com/mapogolions/resilience"
)

type RetryCondition[T any] func(context.Context, resilience.PolicyOutcome[T], int) bool

func RetryOnErrorCondition[T any](ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
	return outcome.Err != nil
}

func NewRetryOnErrorWithDelayCondition[T any](delay func(attempts int) time.Duration) RetryCondition[T] {
	return func(ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
		if RetryOnErrorCondition[T](ctx, outcome, attempts) {
			d := delay(attempts)
			time.Sleep(d)
			return true
		}
		return false
	}
}

func NewRetryPolicy[S any, T any](retryCount int, condition RetryCondition[T]) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var result T
		var err error
		var attempts int
		for {
			result, err = f(ctx, s)
			if attempts >= retryCount {
				return result, err
			}
			attempts++
			outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
			if !condition(ctx, outcome, attempts) {
				return result, err
			}
		}
	}
}
