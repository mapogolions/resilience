package timeout

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

func ExecutePessimistic[T any, R any](
	ctx context.Context,
	f func(context.Context, T) (R, error),
	state T,
	timeout time.Duration,
) (R, error) {
	var defaultValue R
	if ctx.Err() != nil {
		return defaultValue, ctx.Err()
	}
	deadline := time.Now().Add(timeout)
	timeoutCtx, timeoutCancel := context.WithDeadline(ctx, deadline)
	defer timeoutCancel()

	dataCh := func() <-chan result[R] {
		ch := make(chan result[R], 1)
		go func() {
			defer close(ch)
			v, err := f(timeoutCtx, state)
			ch <- result[R]{v, err}
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

func ExecuteOptimistic[T any, R any](
	ctx context.Context,
	f func(context.Context, T) (R, error),
	state T,
	timeout time.Duration,
) (R, error) {
	var defaultValue R
	if ctx.Err() != nil {
		return defaultValue, ctx.Err()
	}
	deadline := time.Now().Add(timeout)
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
