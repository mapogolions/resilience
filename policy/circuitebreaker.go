package policy

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience"
	"github.com/mapogolions/resilience/internal"
)

var ErrCircuitBroken error = errors.New("circuit broken")

type CircuitCommit[T any] func(T, error)
type CircuitFunc[T any] func(CircuitCommit[T]) (T, error)
type CircuitContinuation[T any] func(CircuitFunc[T]) (T, error)
type CircuitBreaker[T any] func() (CircuitContinuation[T], bool)

func NewConsecutiveFailuresCircuitBreaker[T any](
	failureThreshold int,
	breakDuration time.Duration,
	condition OutcomeAcceptanceCondition[T],
) CircuitBreaker[T] {
	circuitBreaker := internal.NewCircuitBreaker[T](failureThreshold, breakDuration, internal.DefaultTimeProvider)
	var commit CircuitCommit[T] = func(result T, err error) {
		if condition(Outcome[T]{Result: result, Err: err}) {
			circuitBreaker.Success()
		} else {
			circuitBreaker.Failure(result, err)
		}
	}
	return func() (CircuitContinuation[T], bool) {
		if circuitBreaker.IsCircuitOpen() {
			return nil, false
		}
		return func(cbf CircuitFunc[T]) (T, error) { return cbf(commit) }, true
	}
}

func NewCircuitBreakerPolicy[S any, T any](circuitBreaker CircuitBreaker[T]) resilience.Policy[S, T] {
	var defaltT T
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		if continuation, ok := circuitBreaker(); ok {
			fn := func(commit CircuitCommit[T]) (T, error) {
				result, err := f(ctx, s)
				commit(result, err)
				return result, err
			}
			return continuation(fn)
		}
		return defaltT, ErrCircuitBroken
	}
}
