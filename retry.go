package resilience

import (
	"context"
	"time"
)

type RetryCondition[T any] func(outcome Outcome[T], retries int) bool
type DelayProvider func(int) time.Duration

func RetryOnError[T any](retryCount int) RetryCondition[T] {
	return func(outcome Outcome[T], retries int) bool {
		return retries < retryCount && outcome.Err != nil
	}
}

func NewRetryPolicy[S, T any](condition RetryCondition[T]) Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var (
			result, zero T
			err          error
			retries      int
		)
		for {
			if err := ctx.Err(); err != nil {
				return zero, err
			}
			result, err = f(ctx, s)
			outcome := Outcome[T]{Result: result, Err: err}
			if !condition(outcome, retries) {
				return result, err
			}
			retries++
		}
	}
}

func NewRetryPolicyWithDelay[S, T any](
	condition RetryCondition[T],
	delayProvider DelayProvider) Policy[S, T] {

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var (
			result, zero T
			err          error
		)
		for retries := 0; ; retries++ {
			if err := ctx.Err(); err != nil {
				return zero, err
			}
			result, err = f(ctx, s)
			outcome := Outcome[T]{Result: result, Err: err}
			if !condition(outcome, retries) {
				return result, err
			}
			timer := time.NewTimer(delayProvider(retries))
			select {
			case <-ctx.Done():
				timer.Stop()
				return result, ctx.Err()
			case <-timer.C:
			}
		}
	}
}

func (rc RetryCondition[T]) Or(conditions ...RetryCondition[T]) RetryCondition[T] {
	return func(outcome Outcome[T], retries int) bool {
		if rc(outcome, retries) {
			return true
		}
		for _, condition := range conditions {
			if condition(outcome, retries) {
				return true
			}
		}
		return false
	}
}

func (rc RetryCondition[T]) And(conditions ...RetryCondition[T]) RetryCondition[T] {
	return func(outcome Outcome[T], retries int) bool {
		if !rc(outcome, retries) {
			return false
		}
		for _, condition := range conditions {
			if !condition(outcome, retries) {
				return false
			}
		}
		return true
	}
}
