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
	concurrencyBarrier := internal.NewBarrier(concurrency)
	queueBarrier := internal.NewBarrier(concurrency + queue)

	return func(ctx context.Context, f func(context.Context, T) (R, error), state T) (R, error) {
		var defaultValue R
		if !queueBarrier.TryWait() {
			return defaultValue, ErrBulkheadRejected
		}
		concurrencyBarrier.Wait()
		value, err := f(ctx, state)
		concurrencyBarrier.Release()
		queueBarrier.Release()
		return value, err
	}
}
