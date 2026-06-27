package policy

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mapogolions/resilience"
)

var ErrDebounced = errors.New("call debounced")

type DebouncePolicyKind int

const (
	DebounceFirst DebouncePolicyKind = 0
)

func NewDebouncePolicy[S any, T any](d time.Duration, kind DebouncePolicyKind) resilience.Policy[S, T] {
	if kind == DebounceFirst {
		return debounceFirstPolicy[S, T](d)
	}
	panic("not supported")
}

func debounceFirstPolicy[S any, T any](d time.Duration) resilience.Policy[S, T] {
	var lastTime time.Time
	m := sync.Mutex{}

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		m.Lock()
		defer func() {
			lastTime = time.Now().Add(d)
			m.Unlock()
		}()
		if time.Now().Before(lastTime) {
			var zero T
			return zero, ErrDebounced
		}
		return f(ctx, s)
	}
}
