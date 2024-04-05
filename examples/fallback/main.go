package main

import (
	"context"

	"github.com/mapogolions/resilience/policy"
)

func main() {
	type S string
	type T []byte

	policy.NewFallbackPolicy[S, T](policy.IdentityFallback[T])

	policy.NewFallbackPolicy[S, T](func(ctx context.Context, o policy.Outcome[T]) (T, error) {
		panic("not implemented")
	})

	policy.NewPanicFallbackPolicy[S, T](policy.IdentityFallback[T])

	policy.NewPanicFallbackPolicy[S, T](func(ctx context.Context, o policy.Outcome[T]) (T, error) {
		panic("not implemented")
	})
}
