package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mapogolions/resilience"
)

func main() {
	f := func(ctx context.Context, state string) (int, error) {
		return len(state), nil
	}
	result, err := resilience.ExecuteOptimistic[string, int](context.Background(), f, "foo", 2*time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
