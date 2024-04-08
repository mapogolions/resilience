package main

import (
	"context"
	"time"

	"github.com/mapogolions/resilience/policy"
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
	policy.NewTimeoutPolicy[S, T](timeout, policy.OptimisticTimeoutPolicy)

	// Pessimistic timeout policy
	policy.NewTimeoutPolicy[S, T](timeout, policy.PessimisticTimeoutPolicy)
}
