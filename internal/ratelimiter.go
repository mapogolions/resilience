package internal

import (
	"sync/atomic"
	"time"
)

type lockFreeRateLimiter struct {
	freeTokens   *atomic.Int64
	tokenGenTime *atomic.Int64
	capacity     int64
	tokenPerUnit time.Duration
	timeProvider timeProvider
}

func NewLockFreeRateLimiter(tokenPerUnit time.Duration, capacity int64, timeProvider timeProvider) lockFreeRateLimiter {
	freeTokens := atomic.Int64{}
	freeTokens.Store(capacity)
	tokenGenTime := atomic.Int64{}
	tokenGenTime.Store(timeProvider.UtcNow().UnixMicro() + tokenPerUnit.Microseconds())
	return lockFreeRateLimiter{
		freeTokens:   &freeTokens,
		tokenGenTime: &tokenGenTime,
		capacity:     capacity,
		tokenPerUnit: tokenPerUnit,
		timeProvider: timeProvider,
	}
}

func (rl lockFreeRateLimiter) Try() (bool, time.Duration) {
	tokenPerUnitMicro := rl.tokenPerUnit.Microseconds()
	for {
		restTokens := rl.freeTokens.Add(-1)
		if restTokens >= 0 {
			return true, 0
		}
		now := rl.timeProvider.UtcNow().UnixMicro()
		curTokenGenTime := rl.tokenGenTime.Load()
		delta := now - curTokenGenTime
		if delta < 0 {
			return false, time.Duration(-delta)
		}
		growth := 1 + delta/tokenPerUnitMicro
		tokens := minInt64(rl.capacity, growth)
		var nextTokenGenTime int64
		if tokens < rl.capacity {
			nextTokenGenTime = curTokenGenTime + tokens*tokenPerUnitMicro
		} else {
			nextTokenGenTime = now + tokenPerUnitMicro
		}
		if rl.tokenGenTime.CompareAndSwap(curTokenGenTime, nextTokenGenTime) {
			// give one token to the winner
			rl.freeTokens.Store(tokens - 1)
			return true, 0
		}
		time.Sleep(0)
	}
}

func minInt64(a int64, b int64) int64 {
	if a <= b {
		return a
	}
	return b
}
