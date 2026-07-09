package main

import (
	"context"
	"time"

	"github.com/mapogolions/resilience"
)

func main() {
	type S string
	type T []byte
	var timeout time.Duration

	// Standard approach
	{
		f := func(context.Context) {}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		f(ctx)
	}

	// Optimistic timeout policy
	resilience.NewTimeoutPolicy[S, T](timeout, resilience.OptimisticTimeoutPolicy)

	// Pessimistic timeout policy
	resilience.NewTimeoutPolicy[S, T](timeout, resilience.PessimisticTimeoutPolicy)
}
