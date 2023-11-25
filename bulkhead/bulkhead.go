package bulkhead

import (
	"context"
	"errors"

	"github.com/mapogolions/resilience/internal"
)

var ErrBulkheadRejected = errors.New("bulkhead rejected")

func Execute[T any, R any](
	ctx context.Context,
	f func(context.Context, T) (R, error),
	state T,
	concurrency int,
	queue int,
) (R, error) {
	var defaultValue R
	concurrencyBarrier := internal.NewBarrier(concurrency)
	queueBarrier := internal.NewBarrier(concurrency + queue)
	if queueBarrier.TryWait() {
		return defaultValue, ErrBulkheadRejected
	}
	concurrencyBarrier.Wait()
	value, err := f(ctx, state)
	concurrencyBarrier.Release()
	queueBarrier.Release()
	return value, err
}
