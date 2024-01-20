package policy

import (
	"context"

	"github.com/mapogolions/resilience"
)

func Combine[S any, T any](outer resilience.Policy[S, T], inner resilience.Policy[S, T]) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		return outer(ctx, func(ctx context.Context, s S) (T, error) {
			return inner(ctx, f, s)
		}, s)
	}
}

func Pipeline[S any, T any](policies ...resilience.Policy[S, T]) resilience.Policy[S, T] {
	if len(policies) == 0 {
		return NewIdentityPolicy[S, T]()
	}
	if len(policies) == 1 {
		return policies[0]
	}
	policy := Combine(policies[0], policies[1])
	// Avoid using the following construct: `return Combine(policy, Pipeline(polices[2:]...))`
	// This helps prevent the addition of redundant policies to the pipeline when the number of policies is even
	rest := policies[2:]
	if len(rest) == 0 {
		return policy
	}
	return Combine(Combine(policies[0], policies[1]), Pipeline(rest...))
}
