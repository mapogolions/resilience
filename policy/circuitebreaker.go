package policy

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience"
	"github.com/mapogolions/resilience/internal"
)

var ErrBrokenCircuite error = errors.New("broken curcuite")

type CircuiteBreakerCondition[T any] func(resilience.PolicyOutcome[T]) bool

func NewCircuiteBreakerPolicy[S any, T any](
	breakDuration time.Duration,
	condition CircuiteBreakerCondition[T],
) resilience.Policy[S, T] {
	circuiteBreaker := internal.NewCircuiteBreaker[T]()
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		var defaultT T
		if err := circuiteBreaker.Before(); err != nil {
			return defaultT, ErrBrokenCircuite
		}
		result, err := f(ctx, s)
		outcome := resilience.PolicyOutcome[T]{Result: result, Err: err}
		if condition(outcome) {
			circuiteBreaker.Success()
		} else {
			circuiteBreaker.Failure(result, err)
		}
		return result, err
	}
}
