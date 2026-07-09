package resilience

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience/internal"
)

var ErrRateLimitRejected = errors.New("rate limit rejected")

type RateLimitCondition func() (bool, time.Duration)

func NewTokenBucketRateLimitCondition(tokenPerUnit time.Duration, capacity int64) RateLimitCondition {
	rateLimiter := internal.NewLockFreeRateLimiter(tokenPerUnit, capacity, internal.DefaultTimeProvider)
	return rateLimiter.Try
}

func (pf PolicyFunc[S, T]) TokenBucketRateLimit(tokenPerUnit time.Duration, capacity int64) PolicyFunc[S, T] {
	return NewTokenBucketRateLimitPolicy[S, T](tokenPerUnit, capacity).Bind(pf)
}

func NewTokenBucketRateLimitPolicy[S any, T any](tokenPerUnit time.Duration, capacity int64) Policy[S, T] {
	shouldLimit := NewTokenBucketRateLimitCondition(tokenPerUnit, capacity)
	return NewRateLimitPolicy[S, T](shouldLimit)
}

func (pf PolicyFunc[S, T]) RateLimit(limitCondition RateLimitCondition) PolicyFunc[S, T] {
	return NewRateLimitPolicy[S, T](limitCondition).Bind(pf)
}

func NewRateLimitPolicy[S any, T any](limitCondition RateLimitCondition) Policy[S, T] {
	var zero T

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		if ok, _ := limitCondition(); !ok {
			return zero, ErrRateLimitRejected
		}
		return f(ctx, s)
	}
}
