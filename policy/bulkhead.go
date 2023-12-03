package policy

import (
	"context"
	"errors"

	"github.com/mapogolions/resilience/internal"
)

var ErrBulkheadRejected = errors.New("bulkhead rejected")

type BulkheadPolicy[S any, T any] struct {
	concurrency        int
	concurrencyLimiter *internal.Semaphore
	queue              int
	queueLimiter       *internal.Semaphore
}

func (p *BulkheadPolicy[S, T]) Apply(ctx context.Context, f func(context.Context, S) (T, error), state S) (T, error) {
	var defaultValue T
	if !p.queueLimiter.TryWait() {
		return defaultValue, ErrBulkheadRejected
	}
	p.concurrencyLimiter.Wait()
	value, err := f(ctx, state)
	p.concurrencyLimiter.Release()
	p.queueLimiter.Release()
	return value, err
}

func NewBulkheadPolicy[S any, T any](concurrency int, queue int) *BulkheadPolicy[S, T] {
	if concurrency < 0 {
		panic("concurrency must be >= 0")
	}
	if queue < 0 {
		panic("queue must be >= 0")
	}
	return &BulkheadPolicy[S, T]{
		concurrency:        concurrency,
		concurrencyLimiter: internal.NewSemaphore(concurrency),
		queue:              queue,
		queueLimiter:       internal.NewSemaphore(concurrency + queue),
	}
}
