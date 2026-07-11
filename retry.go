package resilience

import (
	"context"
	"time"
)

type RetryCondition[T any] func(result T, err error, retries int) bool
type DelayProvider func(int) time.Duration

func RetryOnError[T any](retryCount int) RetryCondition[T] {
	return func(_ T, err error, retries int) bool {
		return retries < retryCount && err != nil
	}
}

func (pf PolicyFunc[S, T]) Retry(condition RetryCondition[T]) PolicyFunc[S, T] {
	return NewRetryPolicy[S, T](condition).Bind(pf)
}

func NewRetryPolicy[S, T any](condition RetryCondition[T]) Policy[S, T] {
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
			if !condition(result, err, retries) {
				return result, err
			}
		}
	}
}

func (pf PolicyFunc[S, T]) RetryWithDelay(
	condition RetryCondition[T],
	delayProvider DelayProvider) PolicyFunc[S, T] {

	return NewRetryPolicyWithDelay[S, T](condition, delayProvider).Bind(pf)
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
			if !condition(result, err, retries) {
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
	return func(result T, err error, retries int) bool {
		if rc(result, err, retries) {
			return true
		}
		for _, condition := range conditions {
			if condition(result, err, retries) {
				return true
			}
		}
		return false
	}
}

func (rc RetryCondition[T]) And(conditions ...RetryCondition[T]) RetryCondition[T] {
	return func(result T, err error, retries int) bool {
		if !rc(result, err, retries) {
			return false
		}
		for _, condition := range conditions {
			if !condition(result, err, retries) {
				return false
			}
		}
		return true
	}
}
