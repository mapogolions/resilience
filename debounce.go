package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrDebounced = errors.New("call debounced")

func (pf PolicyFunc[S, T]) DebounceFirst(d time.Duration) PolicyFunc[S, T] {
	return NewDebounceFirstPolicy[S, T](d).Bind(pf)
}

func NewDebounceFirstPolicy[S any, T any](d time.Duration) Policy[S, T] {
	var nextCallTime time.Time
	m := sync.Mutex{}

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		m.Lock()
		now := time.Now()

		if now.Before(nextCallTime) {
			m.Unlock()
			var zero T
			return zero, ErrDebounced
		}
		nextCallTime = now.Add(d)
		m.Unlock()

		return f(ctx, s)
	}
}
