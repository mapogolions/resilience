package policy

import (
	"context"

	"github.com/mapogolions/resilience"
)

func IdentityFallback[T any](ctx context.Context, outcome resilience.PolicyOutcome[T]) (T, error) {
	return outcome.Result, outcome.Err
}

func NewIdentityPolicy[S any, T any]() resilience.Policy[S, T] {
	return NewFallbackPolicy[S, T](IdentityFallback)
}
