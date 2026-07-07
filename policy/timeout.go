package policy

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience"
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

func NewTimeoutPolicy[S any, T any](timeout time.Duration, kind TimeoutPolicyKind) resilience.Policy[S, T] {
	if kind == OptimisticTimeoutPolicy {
		return optimisticTimeout[S, T](timeout)
	}
	if kind == PessimisticTimeoutPolicy {
		return pessimisticTimeout[S, T](timeout)
	}
	panic("not supported")
}

func pessimisticTimeout[S any, T any](timeout time.Duration) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var zero T
		if ctx.Err() != nil {
			return zero, ctx.Err()
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
			return zero, ErrTimeoutRejected
		case result := <-dataCh:
			if result.Err == nil {
				return result.Value, nil
			}
			if errors.Is(result.Err, context.DeadlineExceeded) && !isInheritParentTimeout(deadline, ctx) {
				return zero, ErrTimeoutRejected
			}
			return zero, result.Err
		}
	}
}

func optimisticTimeout[S any, T any](timeout time.Duration) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var zero T
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		deadline := time.Now().Add(timeout)
		timeoutCtx, timeoutCancel := context.WithDeadline(ctx, deadline)
		defer timeoutCancel()
		value, err := f(timeoutCtx, s)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) && !isInheritParentTimeout(deadline, ctx) {
				return zero, ErrTimeoutRejected
			}
			return zero, err
		}
		return value, nil
	}
}

func isInheritParentTimeout(deadline time.Time, ctx context.Context) bool {
	parentDeadline, ok := ctx.Deadline()
	return ok && parentDeadline.Before(deadline)
}
