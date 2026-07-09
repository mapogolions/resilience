package resilience

import (
	"context"
	"errors"

	"github.com/mapogolions/resilience/internal"
)

var ErrBulkheadRejected = errors.New("bulkhead rejected")

func (pf PolicyFunc[S, T]) Bulkhead(concurrency int, queue int) PolicyFunc[S, T] {
	policy := NewBulkheadPolicy[S, T](concurrency, queue)
	return policy.Bind(pf)
}

func NewBulkheadPolicy[S any, T any](concurrency int, queue int) Policy[S, T] {
	if concurrency < 0 {
		panic("concurrency must be >= 0")
	}
	if queue < 0 {
		panic("queue must be >= 0")
	}
	concurrencyLimiter := internal.NewSemaphore(concurrency)
	queueLimiter := internal.NewSemaphore(concurrency + queue)
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var zero T
		if !queueLimiter.TryWait() {
			return zero, ErrBulkheadRejected
		}
		concurrencyLimiter.Wait()
		value, err := f(ctx, s)
		concurrencyLimiter.Release()
		queueLimiter.Release()
		return value, err
	}
}
