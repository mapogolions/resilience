package policy

type Outcome[T any] struct {
	Result T
	Err    error
}

type OutcomeAcceptanceCondition[T any] func(Outcome[T]) bool
