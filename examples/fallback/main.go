package main

import (
	"context"

	"github.com/mapogolions/resilience"
)

func main() {
	type S string
	type T []byte

	resilience.NewFallbackPolicy[S, T](resilience.IdentityFallback[T])

	resilience.NewFallbackPolicy[S, T](func(_ context.Context, _ T, _ error) (T, error) {
		panic("not implemented")
	})

	resilience.NewPanicFallbackPolicy[S, T](resilience.IdentityFallback[T])

	resilience.NewPanicFallbackPolicy[S, T](func(_ context.Context, _ T, _ error) (T, error) {
		panic("not implemented")
	})
}
