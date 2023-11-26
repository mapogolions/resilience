package bulkhead

import (
	"context"
	"errors"

	"github.com/mapogolions/resilience/internal"
)

var ErrBulkheadRejected = errors.New("bulkhead rejected")

func Execute[T any, R any](
	concurrency int,
	queue int,
) func(context.Context, func(context.Context, T) (R, error), T) (R, error) {
	concurrencyLimiter := internal.NewSemaphore(concurrency)
	queueLimiter := internal.NewSemaphore(concurrency + queue)

	return func(ctx context.Context, f func(context.Context, T) (R, error), state T) (R, error) {
		var defaultValue R
		if !queueLimiter.TryWait() {
			return defaultValue, ErrBulkheadRejected
		}
		concurrencyLimiter.Wait()
		value, err := f(ctx, state)
		concurrencyLimiter.Release()
		queueLimiter.Release()
		return value, err
	}
}
