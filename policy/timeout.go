package policy

import (
	"context"
	"errors"
	"time"
)

var ErrTimeoutRejected = errors.New("rejected by timeout")

type result[T any] struct {
	Value T
	Err   error
}

type TimeoutPolicyKind int

const (
	OptimisticTimeoutPolicy  TimeoutPolicyKind = 0
	PessimisticTimeoutPolicy TimeoutPolicyKind = 1
)

type TimeoutPolicy[S any, T any] func(context.Context, func(context.Context, S) (T, error), S) (T, error)

func NewTimeoutPolicy[S any, T any](timeout time.Duration, kind TimeoutPolicyKind) TimeoutPolicy[S, T] {
	if kind == OptimisticTimeoutPolicy {
		return newOptimisticTimeoutPolicy[S, T](timeout)
	}
	return newPessimisticTimeoutPolicy[S, T](timeout)
}

func newPessimisticTimeoutPolicy[S any, T any](timeout time.Duration) TimeoutPolicy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var defaultValue T
		if ctx.Err() != nil {
			return defaultValue, ctx.Err()
		}
		deadline := time.Now().Add(timeout)
		timeoutCtx, timeoutCancel := context.WithDeadline(ctx, deadline)
		defer timeoutCancel()

		dataCh := func() <-chan result[T] {
			ch := make(chan result[T], 1)
			go func() {
				defer close(ch)
				v, err := f(timeoutCtx, s)
				ch <- result[T]{v, err}
			}()
			return ch
		}()

		select {
		case <-timeoutCtx.Done():
			return defaultValue, ErrTimeoutRejected
		case result := <-dataCh:
			if result.Err == nil {
				return result.Value, nil
			}
			if errors.Is(result.Err, context.DeadlineExceeded) && !isInheritParentTimeout(deadline, ctx) {
				return defaultValue, ErrTimeoutRejected
			}
			return defaultValue, result.Err
		}
	}
}

func newOptimisticTimeoutPolicy[S any, T any](timeout time.Duration) TimeoutPolicy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var defaultValue T
		if ctx.Err() != nil {
			return defaultValue, ctx.Err()
		}
		deadline := time.Now().Add(timeout)
		timeoutCtx, timeoutCancel := context.WithDeadline(ctx, deadline)
		defer timeoutCancel()
		value, err := f(timeoutCtx, s)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) && !isInheritParentTimeout(deadline, ctx) {
				return defaultValue, ErrTimeoutRejected
			}
			return defaultValue, err
		}
		return value, nil
	}
}

func isInheritParentTimeout(deadline time.Time, ctx context.Context) bool {
	parentDeadline, ok := ctx.Deadline()
	return ok && parentDeadline.Before(deadline)
}
