package policy

import (
	"context"
	"time"

	"github.com/mapogolions/resilience"
)

type RetryCondition[T any] func(ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool

func RetryOnErrorCondition[T any](_ context.Context, outcome resilience.PolicyOutcome[T], _ int) bool {
	return outcome.Err != nil
}

func NewRetryCountOnErrorCondition[T any](retryCount int) RetryCondition[T] {
	return func(ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
		return attempts < retryCount && RetryOnErrorCondition[T](ctx, outcome, attempts)
	}
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

func (rc RetryCondition[T]) Or(conditions ...RetryCondition[T]) RetryCondition[T] {
	return func(ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
		if rc(ctx, outcome, attempts) {
			return true
		}
		for _, condition := range conditions {
			if condition(ctx, outcome, attempts) {
				return true
			}
		}
		return false
	}
}

func (rc RetryCondition[T]) And(conditions ...RetryCondition[T]) RetryCondition[T] {
	return func(ctx context.Context, outcome resilience.PolicyOutcome[T], attempts int) bool {
		if !rc(ctx, outcome, attempts) {
			return false
		}
		for _, condition := range conditions {
			if !condition(ctx, outcome, attempts) {
				return false
			}
		}
		return true
	}
}
