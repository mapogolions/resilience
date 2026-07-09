package main

import (
	"context"

	"github.com/mapogolions/resilience"
)

func main() {
	type S string
	type T []byte

	resilience.NewFallbackPolicy[S, T](resilience.IdentityFallback[T])

	resilience.NewFallbackPolicy[S, T](func(ctx context.Context, o resilience.Outcome[T]) (T, error) {
		panic("not implemented")
	})

	resilience.NewPanicFallbackPolicy[S, T](resilience.IdentityFallback[T])

	resilience.NewPanicFallbackPolicy[S, T](func(ctx context.Context, o resilience.Outcome[T]) (T, error) {
		panic("not implemented")
	})
}
