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
	var (
		zero, result T
		err          error = ErrDebounced
		timer        *time.Timer
		lastCall     call[S]
		m            = sync.Mutex{}
	)

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		m.Lock()
		defer m.Unlock()

		lastCall = call[S]{ctx: ctx, s: s}

		r, e := result, err
		if timer != nil {
			timer.Reset(d)
			return r, e
		}

		result, err = zero, ErrDebounced
		timer = time.NewTimer(d)

		go func() {
			<-timer.C
			m.Lock()
			defer m.Unlock()
			result, err = f(lastCall.ctx, lastCall.s)
			timer = nil
		}()

		return r, e
	}
}

type call[S any] struct {
	ctx context.Context
	s   S
}
