package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mapogolions/resilience/policy"
)

var ErrSlowResponse = errors.New("slow response")

func main() {
	// treat certain results as errors or encapsulate error message
	var fallback policy.FallbackFunc[time.Duration] = func(_ context.Context, o policy.Outcome[time.Duration]) (time.Duration, error) {
		if o.Err == nil && o.Result > 200*time.Millisecond {
			return 0, ErrSlowResponse
		}
		return o.Result, fmt.Errorf("ping error: %v", o.Err)
	}
	policy := policy.NewFallbackPolicy[string, time.Duration](fallback)
	r, err := policy(context.Background(), ping, "https://github.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)
}

func ping(ctx context.Context, url string) (time.Duration, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	return time.Since(start), nil
}
