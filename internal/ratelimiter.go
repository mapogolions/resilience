package internal

import (
	"runtime"
	"sync/atomic"
	"time"
)

type lockFreeTokenBucketRateLimiter struct {
	freeTokens   *atomic.Int64
	tokenGenTime *atomic.Int64
	capacity     int64
	tokenPerUnit time.Duration
	timeProvider timeProvider
}

func NewLockFreeTokenBucketRateLimiter(
	tokenPerUnit time.Duration,
	capacity int64,
	timeProvider timeProvider) lockFreeTokenBucketRateLimiter {

	freeTokens := atomic.Int64{}
	freeTokens.Store(capacity)

	tokenGenTime := atomic.Int64{}
	tokenGenTime.Store(timeProvider.UtcNow().UnixNano() + tokenPerUnit.Nanoseconds())

	return lockFreeTokenBucketRateLimiter{
		freeTokens:   &freeTokens,
		tokenGenTime: &tokenGenTime,
		capacity:     capacity,
		tokenPerUnit: tokenPerUnit,
		timeProvider: timeProvider,
	}
}

func (rl lockFreeTokenBucketRateLimiter) Try() (bool, time.Duration) {
	tokenPerUnitNano := rl.tokenPerUnit.Nanoseconds()
	for {
		restTokens := rl.freeTokens.Add(-1)
		if restTokens >= 0 {
			return true, 0
		}
		now := rl.timeProvider.UtcNow().UnixNano()
		curTokenGenTime := rl.tokenGenTime.Load()
		delta := now - curTokenGenTime
		if delta < 0 {
			return false, time.Duration(-delta)
		}
		growth := 1 + delta/tokenPerUnitNano
		tokens := minInt64(rl.capacity, growth)
		var nextTokenGenTime int64
		if tokens < rl.capacity {
			nextTokenGenTime = curTokenGenTime + tokens*tokenPerUnitNano
		} else {
			nextTokenGenTime = now + tokenPerUnitNano
		}
		if rl.tokenGenTime.CompareAndSwap(curTokenGenTime, nextTokenGenTime) {
			// give one token to the winner
			//
			// The implementation contains a bug. A race condition may be possible!!!
			// This is acceptable for educational purposes.
			rl.freeTokens.Store(tokens - 1)
			return true, 0
		}
		runtime.Gosched()
	}
}

func minInt64(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
