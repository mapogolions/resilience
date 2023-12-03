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

type TimeoutPolicy[S any, T any] struct {
	timeout time.Duration
	kind    TimeoutPolicyKind
}

func NewTimeoutPolicy[S any, T any](timeout time.Duration, kind TimeoutPolicyKind) *TimeoutPolicy[S, T] {
	return &TimeoutPolicy[S, T]{timeout: timeout, kind: kind}
}

func (p *TimeoutPolicy[S, T]) Apply(ctx context.Context, f func(context.Context, S) (T, error), state S) (T, error) {
	if p.kind == OptimisticTimeoutPolicy {
		return p.applyOptimistic(ctx, f, state)
	}
	return p.applyPessimistic(ctx, f, state)
}

func (p *TimeoutPolicy[S, T]) applyPessimistic(ctx context.Context, f func(context.Context, S) (T, error), state S) (T, error) {
	if p.kind == OptimisticTimeoutPolicy {
		panic("should be pessimistic")
	}
	var defaultValue T
	if ctx.Err() != nil {
		return defaultValue, ctx.Err()
	}
	deadline := time.Now().Add(p.timeout)
	timeoutCtx, timeoutCancel := context.WithDeadline(ctx, deadline)
	defer timeoutCancel()

	dataCh := func() <-chan result[T] {
		ch := make(chan result[T], 1)
		go func() {
			defer close(ch)
			v, err := f(timeoutCtx, state)
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

func (p *TimeoutPolicy[S, T]) applyOptimistic(ctx context.Context, f func(context.Context, S) (T, error), state S) (T, error) {
	if p.kind == PessimisticTimeoutPolicy {
		panic("should be optimistic")
	}
	var defaultValue T
	if ctx.Err() != nil {
		return defaultValue, ctx.Err()
	}
	deadline := time.Now().Add(p.timeout)
	timeoutCtx, timeoutCancel := context.WithDeadline(ctx, deadline)
	defer timeoutCancel()
	value, err := f(timeoutCtx, state)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) && !isInheritParentTimeout(deadline, ctx) {
			return defaultValue, ErrTimeoutRejected
		}
		return defaultValue, err
	}
	return value, nil
}

func isInheritParentTimeout(deadline time.Time, ctx context.Context) bool {
	parentDeadline, ok := ctx.Deadline()
	return ok && parentDeadline.Before(deadline)
}
