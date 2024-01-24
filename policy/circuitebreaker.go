package policy

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience"
	"github.com/mapogolions/resilience/internal"
)

var ErrBrokenCircuit error = errors.New("broken curcuite")

type CircuitBreakCondition[T any] func(resilience.PolicyOutcome[T]) bool
type CircuitCommit[T any] func(resilience.PolicyOutcome[T])
type CircuitFunc[T any] func(CircuitCommit[T]) (T, error)
type CircuitContinuation[T any] func(CircuitFunc[T]) (T, error)
type CircuitBreaker[T any] func() (CircuitContinuation[T], bool)

func NewConsecutiveFailuresCircuitBreaker[T any](
	consecutiveFailures int,
	breakDuration time.Duration,
	condition CircuitBreakCondition[T],
) CircuitBreaker[T] {
	circuitBreaker := internal.NewCircuitBreaker[T](consecutiveFailures, breakDuration, internal.DefaultTimeProvider)
	var commit CircuitCommit[T] = func(outcome resilience.PolicyOutcome[T]) {
		if condition(outcome) {
			circuitBreaker.Success()
		} else {
			circuitBreaker.Failure(outcome.Result, outcome.Err)
		}
	}
	return func() (CircuitContinuation[T], bool) {
		if circuitBreaker.Before() {
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
				outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
				commit(outcome)
				return result, err
			}
			return continuation(fn)
		}
		return defaltT, ErrBrokenCircuit
	}
}
