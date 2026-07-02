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
		return debounceFirst[S, T](d)
	}
	panic("not supported")
}

func debounceFirst[S any, T any](d time.Duration) resilience.Policy[S, T] {
	var lastCallTime time.Time
	m := sync.Mutex{}

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		m.Lock()
		now := time.Now()

		if now.Before(lastCallTime) {
			m.Unlock()
			var zero T
			return zero, ErrDebounced
		}
		lastCallTime = now.Add(d)
		m.Unlock()

		return f(ctx, s)
	}

}
