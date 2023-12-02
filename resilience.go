package resilience

type DelegateResult[T any] struct {
	Result T
	Err    error
}

type ResultPredicate[T any] func(T) bool
type ResultPredicates[T any] []ResultPredicate[T]

func (predicates ResultPredicates[T]) AnyMatch(result T) bool {
	for _, pred := range predicates {
		if pred(result) {
			return true
		}
	}
	return false
}

type ErrorPredicate func(error) bool
type ErrorPredicates []ErrorPredicate

func (predicates ErrorPredicates) AnyMatch(err error) bool {
	for _, pred := range predicates {
		if pred(err) {
			return true
		}
	}
	return false
}
