package main

import (
	"context"
	"time"

	"github.com/mapogolions/resilience/policy"
)

func main() {
	type S string
	type T []byte
	var N int

	policy.NewRetryPolicy[S, T](
		policy.NewRetryCountOnErrorCondition[T](10),
	)

	policy.NewRetryPolicy[S, T](
		policy.NewRetryCountOnErrorWithDelayCondition[T](
			N,
			func(i int) time.Duration {
				return time.Duration(i) * time.Second
			},
		),
	)

	policy.NewRetryPolicy[S, T](func(ctx context.Context, outcome policy.Outcome[T], retries int) bool {
		panic("not implemented")
	})
}
