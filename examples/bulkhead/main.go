package main

import "github.com/mapogolions/resilience"

func main() {
	type S string
	type T []byte
	var CONCURRENCY, QUEUE int

	resilience.NewBulkheadPolicy[S, T](CONCURRENCY, QUEUE)
}
