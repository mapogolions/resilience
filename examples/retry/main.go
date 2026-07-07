package main

import (
	"github.com/mapogolions/resilience/policy"
)

func main() {
	type S string
	type T []byte

	policy.NewRetryPolicy[S, T](
		policy.RetryOnError[T](10),
	)
}
