package policy

import (
	"context"
	"time"

	"github.com/mapogolions/resilience"
)

type RetryCondition[T any] func(ctx context.Context, outcome Outcome[T], retries int) bool

func NewRetryCountOnErrorCondition[T any](retryCount int) RetryCondition[T] {
	return func(ctx context.Context, outcome Outcome[T], retries int) bool {
		return retries < retryCount && outcome.Err != nil
	}
}

func NewRetryCountOnErrorWithDelayCondition[T any](retryCount int, delayProvider func(int) time.Duration) RetryCondition[T] {
	retryCountOnError := NewRetryCountOnErrorCondition[T](retryCount)
	return func(ctx context.Context, outcome Outcome[T], retries int) bool {
		if retryCountOnError(ctx, outcome, retries) {
			defer time.Sleep(delayProvider(retries))
			return true
		}
		return false
	}
}

func NewRetryPolicy[S any, T any](shouldRetry RetryCondition[T]) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var result T
		var err error
		var retries int
		for {
			result, err = f(ctx, s)
			outcome := Outcome[T]{Result: result, Err: err}
			if !shouldRetry(ctx, outcome, retries) {
				return result, err
			}
			retries++
		}
	}
}

func (rc RetryCondition[T]) Or(conditions ...RetryCondition[T]) RetryCondition[T] {
	return func(ctx context.Context, outcome Outcome[T], retries int) bool {
		if rc(ctx, outcome, retries) {
			return true
		}
		for _, condition := range conditions {
			if condition(ctx, outcome, retries) {
				return true
			}
		}
		return false
	}
}

func (rc RetryCondition[T]) And(conditions ...RetryCondition[T]) RetryCondition[T] {
	return func(ctx context.Context, outcome Outcome[T], retries int) bool {
		if !rc(ctx, outcome, retries) {
			return false
		}
		for _, condition := range conditions {
			if !condition(ctx, outcome, retries) {
				return false
			}
		}
		return true
	}
}
