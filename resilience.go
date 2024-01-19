package resilience

import "context"

type Policy[S any, T any] func(context.Context, func(context.Context, S) (T, error), S) (T, error)

type PolicyOutcome[T any] struct {
	Result T
	Err    error
}

type ResultPredicate[T any] func(T) bool
type ResultPredicates[T any] []ResultPredicate[T]

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

func (predicates ErrorPredicates) Any(err error) bool {
	for _, pred := range predicates {
		if pred(err) {
			return true
		}
	}
	return false
}
