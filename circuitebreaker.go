package resilience

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience/internal"
)

var ErrCircuitBroken error = errors.New("circuit broken")

type CircuitCommit[T any] func(T, error)
type CircuitBreaker[T any] func() (CircuitCommit[T], bool)

func ConsecutiveFailuresCircuitBreaker[T any](
	threshold int,
	breakDuration time.Duration,
	p func(T, error) bool,
) CircuitBreaker[T] {

	circuitBreaker := internal.NewCircuitBreaker[T](
		threshold,
		breakDuration,
		internal.DefaultTimeProvider)

	var commit CircuitCommit[T] = func(result T, err error) {
		if p(result, err) {
			circuitBreaker.Success()
			return
		}
		circuitBreaker.Failure(result, err)
	}

	return func() (CircuitCommit[T], bool) {
		if circuitBreaker.TryAcquire() {
			return commit, true
		}
		return nil, false
	}
}

func NewCircuitBreakerPolicy[S any, T any](cb CircuitBreaker[T]) Policy[S, T] {
	var zero T

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		commit, ok := cb()
		if !ok {
			return zero, ErrCircuitBroken
		}
		result, err := f(ctx, s)
		commit(result, err)
		return result, err
	}
}
