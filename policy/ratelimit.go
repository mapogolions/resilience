package policy

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience"
	"github.com/mapogolions/resilience/internal"
)

var ErrRateLimitRejected = errors.New("rate limit rejected")

type RateLimiter func() (bool, time.Duration)

func NewRateLimitPolicy[S any, T any](tokenPerUnit time.Duration, capacity int64) resilience.Policy[S, T] {
	rateLimiter := internal.NewLockFreeRateLimiter(tokenPerUnit, capacity, internal.DefaultTimeProvider)
	return NewRateLimitPolicyWith[S, T](rateLimiter.Try)
}

func NewRateLimitPolicyWith[S any, T any](rateLimiter RateLimiter) resilience.Policy[S, T] {
	var defaultT T
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		if ok, _ := rateLimiter(); !ok {
			return defaultT, ErrRateLimitRejected
		}
		return f(ctx, s)
	}
}
