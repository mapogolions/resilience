package policy

import (
	"context"
	"errors"

	"github.com/mapogolions/resilience"
	"github.com/mapogolions/resilience/internal"
)

var ErrBulkheadRejected = errors.New("bulkhead rejected")

func NewBulkheadPolicy[S any, T any](concurrency int, queue int) resilience.Policy[S, T] {
	if concurrency < 1 {
		panic("concurrency must be > 0")
	}
	if queue < 0 {
		panic("queue must be >= 0")
	}
	concurrencyLimiter := internal.NewSemaphore(concurrency)
	queueLimiter := internal.NewSemaphore(concurrency + queue)
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var defaultValue T
		if !queueLimiter.TryWait() {
			return defaultValue, ErrBulkheadRejected
		}
		concurrencyLimiter.Wait()
		value, err := f(ctx, s)
		concurrencyLimiter.Release()
		queueLimiter.Release()
		return value, err
	}
}
