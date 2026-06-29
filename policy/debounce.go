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
	DebounceLast  DebouncePolicyKind = 1
)

func NewDebouncePolicy[S any, T any](d time.Duration, kind DebouncePolicyKind) resilience.Policy[S, T] {
	if kind == DebounceFirst {
		return debounceFirst[S, T](d)
	}
	if kind == DebounceLast {
		return debounceLast[S, T](d)
	}
	panic("not supported")
}

func debounceFirst[S any, T any](d time.Duration) resilience.Policy[S, T] {
	var lastCallTime time.Time
	m := sync.Mutex{}

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		m.Lock()
		defer func() {
			lastCallTime = time.Now().Add(d)
			m.Unlock()
		}()
		if time.Now().Before(lastCallTime) {
			var zero T
			return zero, ErrDebounced
		}
		return f(ctx, s)
	}
}

func debounceLast[S any, T any](d time.Duration) resilience.Policy[S, T] {
	var result T
	var err error
	m := sync.Mutex{}
	once := sync.Once{}

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		m.Lock()
		defer m.Unlock()

		once.Do(func() {
			timer := time.NewTimer(d)

			go func() {
				defer func() {
					timer.Stop()
					once = sync.Once{}
					m.Unlock()
				}()

				select {
				case <-timer.C:
					m.Lock()
					result, err = f(ctx, s)
					return
				case <-ctx.Done():
					m.Lock()
					err = ctx.Err()
					return
				}
			}()
		})
		return result, err
	}
}
