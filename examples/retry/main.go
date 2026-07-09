package main

import (
	"github.com/mapogolions/resilience"
)

func main() {
	type S string
	type T []byte

	resilience.NewRetryPolicy[S, T](
		resilience.RetryOnError[T](10),
	)
}
