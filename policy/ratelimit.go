package policy

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var ErrRateLimitRejected = errors.New("rate limit rejected")

type RateLimitPolicy[S any, T any] func(context.Context, func(context.Context, S) (T, error), S) (T, error)
type RateLimiter func() (bool, time.Duration)

func NewRateLimitPolicy[S any, T any](tokenPerUnit time.Duration, capacity int64) RateLimitPolicy[S, T] {
	rateLimiter := newRateLimiter(tokenPerUnit, capacity)
	return NewRateLimitPolicyWith[S, T](rateLimiter)
}

func NewRateLimitPolicyWith[S any, T any](rateLimiter RateLimiter) RateLimitPolicy[S, T] {
	var defaultT T
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		if ok, _ := rateLimiter(); ok {
			return defaultT, ErrRateLimitRejected
		}
		return f(ctx, s)
	}
}

func newRateLimiter(tokenPerUnit time.Duration, capacity int64) RateLimiter {
	tokenPerUnitMicrosec := tokenPerUnit.Microseconds()
	var freeTokens atomic.Int64
	freeTokens.Store(capacity)

	var tokenGenTime atomic.Int64
	tokenGenTime.Store(time.Now().UnixMicro() + tokenPerUnitMicrosec)

	return func() (bool, time.Duration) {
		for {
			restTokens := freeTokens.Add(-1)
			if restTokens >= 0 {
				return true, 0
			}
			now := time.Now().UnixMicro()
			curTokenGenTime := tokenGenTime.Load()
			delta := now - curTokenGenTime
			if delta < 0 {
				return false, time.Duration(-delta)
			}
			growth := 1 + delta/tokenPerUnitMicrosec
			tokens := minInt64(capacity, growth)
			var nextTokenGenTime int64
			if tokens < capacity {
				nextTokenGenTime = curTokenGenTime + tokens + tokenPerUnitMicrosec
			} else {
				nextTokenGenTime = now + tokenPerUnitMicrosec
			}
			if tokenGenTime.CompareAndSwap(curTokenGenTime, nextTokenGenTime) {
				// give one token to the winner
				freeTokens.Store(tokens - 1)
				return true, 0
			}
			time.Sleep(0)
		}
	}
}

func minInt64(a int64, b int64) int64 {
	if a <= b {
		return a
	}
	return b
}
