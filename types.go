package resilience

import "context"

type Policy[S, T any] func(context.Context, func(context.Context, S) (T, error), S) (T, error)
type PolicyFunc[S, T any] func(context.Context, S) (T, error)

func (p Policy[S, T]) Bind(f PolicyFunc[S, T]) PolicyFunc[S, T] {
	return func(ctx context.Context, s S) (T, error) {
		return p(ctx, f, s)
	}
}

type Outcome[T any] struct {
	Result T
	Err    error
}

type OutcomeAcceptanceCondition[T any] func(Outcome[T]) bool

func RejectOnError[T any](outcome Outcome[T]) bool {
	return outcome.Err == nil
}
