package main

import "github.com/mapogolions/resilience/policy"

func main() {
	type S string
	type T []byte

	policy.NewBulkheadPolicy[S, T](2, 5)
}
