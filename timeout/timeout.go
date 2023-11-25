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
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, timeout)
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
	case result := <-dataCh:
		if result.Err == nil {
			return result.Value, nil
		}
		if errors.Is(result.Err, context.DeadlineExceeded) && exitByTimeout(ctx, timeoutCtx) {
			return defaultValue, ErrTimeoutRejected
		}
		return defaultValue, result.Err
	case <-timeoutCtx.Done():
		return defaultValue, ErrTimeoutRejected
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
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, timeout)
	defer timeoutCancel()
	value, err := f(timeoutCtx, state)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) && exitByTimeout(ctx, timeoutCtx) {
			return defaultValue, ErrTimeoutRejected
		}
		return defaultValue, err
	}
	return value, nil
}

func exitByTimeout(ctx context.Context, timeoutCtx context.Context) bool {
	timeoutDeadline, ok := timeoutCtx.Deadline()
	if !ok {
		panic("Timeout context should have deadline")
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return true
	}
	return timeoutDeadline.Before(deadline) || timeoutDeadline.Equal(deadline)
}
