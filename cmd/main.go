package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mapogolions/resilience/timeout"
)

func main() {
	f := func(ctx context.Context, state string) (int, error) {
		return len(state), nil
	}
	result, err := timeout.ExecuteOptimistic[string, int](context.Background(), f, "foo", 2*time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
