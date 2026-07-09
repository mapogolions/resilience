package internal

import (
	"context"
)

// BoundedSemaphore is a semaphore with a strict upper bound on the number
// of concurrently acquired slots (similar to Python's
// threading.BoundedSemaphore).
//
// Wait supports cancellation via ctx.
//
// Calling Release without a matching Wait is a programming error and may
// result in a panic or deadlock.

type boundedSemaphore struct {
	slots chan struct{}
}

func NewBoundedSemaphore(threshold int) *boundedSemaphore {
	if threshold < 0 {
		panic("semaphore initial value must be >= 0")
	}
	return &boundedSemaphore{slots: make(chan struct{}, threshold)}
}

func (s *boundedSemaphore) TryWait() bool {
	select {
	case s.slots <- struct{}{}:
		return true
	default:
	}
	return false
}

func (s *boundedSemaphore) Wait(ctx context.Context) error {
	select {
	case s.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *boundedSemaphore) Release() {
	select {
	case <-s.slots:
	default:
		panic("sem: release without matching wait")
	}
}
