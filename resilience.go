package resilience

import (
	"context"

	"github.com/mapogolions/resilience/internal"
)

type Policy[S any, T any] func(context.Context, func(context.Context, S) (T, error), S) (T, error)

type PolicyOutcome[T any] struct {
	Result T
	Err    error
	Dict   map[string]interface{}
}

func (outcome PolicyOutcome[T]) RetryAttempts() int {
	return outcome.Dict[internal.RetryAttemptsKey].(int)
}

type ResultPredicate[T any] func(T) bool
type ResultPredicates[T any] []ResultPredicate[T]

func NewResultPredicates[T any](predicates ...ResultPredicate[T]) ResultPredicates[T] {
	return predicates
}

func (predicates ResultPredicates[T]) Any(result T) bool {
	for _, pred := range predicates {
		if pred(result) {
			return true
		}
	}
	return false
}

type ErrorPredicate func(error) bool
type ErrorPredicates []ErrorPredicate

func NewErrorPredicates(predicats ...ErrorPredicate) ErrorPredicates {
	return predicats
}

func (predicates ErrorPredicates) Any(err error) bool {
	for _, pred := range predicates {
		if pred(err) {
			return true
		}
	}
	return false
}
