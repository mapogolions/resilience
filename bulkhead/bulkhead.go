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
	concurrencyLimit := internal.NewConcurrencyLimiter(concurrency)
	queueLimit := internal.NewConcurrencyLimiter(concurrency + queue)

	return func(ctx context.Context, f func(context.Context, T) (R, error), state T) (R, error) {
		var defaultValue R
		if !queueLimit.TryWait() {
			return defaultValue, ErrBulkheadRejected
		}
		concurrencyLimit.Wait()
		value, err := f(ctx, state)
		concurrencyLimit.Release()
		queueLimit.Release()
		return value, err
	}
}
