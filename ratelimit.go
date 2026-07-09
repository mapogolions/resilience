package resilience

import (
	"context"
	"errors"
	"time"

	"github.com/mapogolions/resilience/internal"
)

var ErrRateLimitRejected = errors.New("rate limit rejected")

type RateLimit func() (bool, time.Duration)

func LockFreeTokenBucketRateLimit(tokenPerUnit time.Duration, capacity int64) RateLimit {
	rateLimiter := internal.NewLockFreeTokenBucketRateLimiter(tokenPerUnit, capacity, internal.DefaultTimeProvider)
	return rateLimiter.Try
}

func (pf PolicyFunc[S, T]) RateLimit(limitCondition RateLimit) PolicyFunc[S, T] {
	return NewRateLimitPolicy[S, T](limitCondition).Bind(pf)
}

func NewRateLimitPolicy[S any, T any](rateLimit RateLimit) Policy[S, T] {
	var zero T

	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		if ok, _ := rateLimit(); !ok {
			return zero, ErrRateLimitRejected
		}
		return f(ctx, s)
	}
}
