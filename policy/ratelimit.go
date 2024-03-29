package policy

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience"
	"github.com/mapogolions/resilience/internal"
)

var ErrRateLimitRejected = errors.New("rate limit rejected")

type RateLimitCondition func() (bool, time.Duration)

func NewTokenBucketRateLimitCondition(tokenPerUnit time.Duration, capacity int64) RateLimitCondition {
	rateLimiter := internal.NewLockFreeRateLimiter(tokenPerUnit, capacity, internal.DefaultTimeProvider)
	return rateLimiter.Try
}

func NewTokenBucketRateLimitPolicy[S any, T any](tokenPerUnit time.Duration, capacity int64) resilience.Policy[S, T] {
	shouldLimit := NewTokenBucketRateLimitCondition(tokenPerUnit, capacity)
	return NewRateLimitPolicy[S, T](shouldLimit)
}

func NewRateLimitPolicy[S any, T any](shouldLimit RateLimitCondition) resilience.Policy[S, T] {
	var defaultT T
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		if ok, _ := shouldLimit(); !ok {
			return defaultT, ErrRateLimitRejected
		}
		return f(ctx, s)
	}
}
