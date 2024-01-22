package policy

import (
	"context"
	"time"

	"github.com/mapogolions/resilience"
)

type RetryCondition[T any] func(ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool

func (rc RetryCondition[T]) Or(conditions ...RetryCondition[T]) RetryCondition[T] {
	if len(conditions) == 0 {
		return rc
	}
	var condition RetryCondition[T] = func(ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
		return rc(ctx, outcome, attempts) || conditions[0](ctx, outcome, attempts)
	}
	return condition.Or(conditions[1:]...)
}

func (rc RetryCondition[T]) And(conditions ...RetryCondition[T]) RetryCondition[T] {
	if len(conditions) == 0 {
		return rc
	}
	var condition RetryCondition[T] = func(ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
		return rc(ctx, outcome, attempts) && conditions[0](ctx, outcome, attempts)
	}
	return condition.And(conditions[1:]...)
}

func RetryOnErrorCondition[T any](_ context.Context, outcome resilience.PolicyOutcome[T], _ int) bool {
	return outcome.Err != nil
}

func NewRetryCountOnErrorCondition[T any](retryCount int) RetryCondition[T] {
	var condition RetryCondition[T] = func(_ context.Context, _ resilience.PolicyOutcome[T], attempts int) bool {
		return attempts < retryCount
	}
	return condition.And(RetryOnErrorCondition[T])
}

func NewRetryCountOnErrorWithDelayCondition[T any](retryCount int, delayProvider func(int) time.Duration) RetryCondition[T] {
	retryCountOnError := NewRetryCountOnErrorCondition[T](retryCount)
	return func(ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
		if retryCountOnError(ctx, outcome, attempts) {
			defer time.Sleep(delayProvider(attempts))
			return true
		}
		return false
	}
}

func NewRetryPolicy[S any, T any](shouldRetry RetryCondition[T]) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var result T
		var err error
		var attempts int
		for {
			result, err = f(ctx, s)
			outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
			if !shouldRetry(ctx, outcome, attempts) {
				return result, err
			}
			attempts++
		}
	}
}
