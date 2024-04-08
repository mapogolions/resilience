package main

import "github.com/mapogolions/resilience/policy"

func main() {
	type S string
	type T []byte
	var CONCURRENCY, QUEUE int

	policy.NewBulkheadPolicy[S, T](CONCURRENCY, QUEUE)
}
