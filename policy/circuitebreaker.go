package policy

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience"
	"github.com/mapogolions/resilience/internal"
)

var ErrBrokenCircuite error = errors.New("broken curcuite")

type CircuiteCondition[T any] func(resilience.PolicyOutcome[T]) bool
type CircuiteCommit[T any] func(resilience.PolicyOutcome[T])
type CircuiteFunc[T any] func(CircuiteCommit[T]) (T, error)
type CircuiteContinuation[T any] func(CircuiteFunc[T]) (T, error)
type CircuiteBreaker[T any] func() (CircuiteContinuation[T], bool)

func NewConsecutiveFailuresCircuiteBreaker[T any](
	consecutiveFailures int,
	breakDuration time.Duration,
	condition CircuiteCondition[T],
) CircuiteBreaker[T] {
	circuiteBreaker := internal.NewCircuiteBreaker[T](consecutiveFailures, breakDuration, internal.DefaultTimeProvider)
	var commit CircuiteCommit[T] = func(outcome resilience.PolicyOutcome[T]) {
		if condition(outcome) {
			circuiteBreaker.Success()
		} else {
			circuiteBreaker.Failure(outcome.Result, outcome.Err)
		}
	}
	return func() (CircuiteContinuation[T], bool) {
		if circuiteBreaker.Before() {
			return nil, false
		}
		return func(cbf CircuiteFunc[T]) (T, error) { return cbf(commit) }, true
	}
}

func NewCircuiteBreakerPolicy[S any, T any](circuiteBreaker CircuiteBreaker[T]) resilience.Policy[S, T] {
	var defaltT T
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		if continuation, ok := circuiteBreaker(); ok {
			fn := func(commit CircuiteCommit[T]) (T, error) {
				result, err := f(ctx, s)
				outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
				commit(outcome)
				return result, err
			}
			return continuation(fn)
		}
		return defaltT, ErrBrokenCircuite
	}
}
