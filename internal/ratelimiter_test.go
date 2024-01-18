package internal

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	t.Run("should not increase free tokens beyond capacity", func(t *testing.T) {
		// Arrange
		utcNow := time.Now().UTC()
		timeProvider := fakeTimeProvider{utcNow}
		rateLimiter := NewLockFreeRateLimiter(1*time.Second, 5, &timeProvider)

		// Act
		exhaustFreeTokens(rateLimiter)
		timeProvider.Advance(100 * time.Second)
		ok, _ := rateLimiter.Try()

		// Assert
		if !ok || rateLimiter.freeTokens.Load() != 4 {
			t.Fail()
		}
	})

	t.Run("should cacl next token generation time correctly when there is a fraction", func(t *testing.T) {
		// Arrange
		utcNow := time.Now().UTC()
		timeProvider := fakeTimeProvider{utcNow}
		rateLimiter := NewLockFreeRateLimiter(1*time.Second, 30, &timeProvider)

		// Act
		exhaustFreeTokens(rateLimiter)
		timeProvider.Advance(10 * time.Second)
		timeProvider.Advance(500 * time.Millisecond)
		ok, _ := rateLimiter.Try()

		// Assert
		if !ok || rateLimiter.freeTokens.Load() != 9 {
			t.Fail()
		}
		if rateLimiter.tokenGenTime.Load() != utcNow.Add(11*time.Second).UnixMicro() {
			t.Fail()
		}
	})

	t.Run("should increase free tokens and cacl next token generation time", func(t *testing.T) {
		// Arrange
		utcNow := time.Now().UTC()
		timeProvider := fakeTimeProvider{utcNow}
		rateLimiter := NewLockFreeRateLimiter(1*time.Second, 30, &timeProvider)

		// Act
		exhaustFreeTokens(rateLimiter)
		timeProvider.Advance(10 * time.Second)
		ok, _ := rateLimiter.Try()

		// Assert
		if !ok || rateLimiter.freeTokens.Load() != 9 {
			t.Fail()
		}
		if rateLimiter.tokenGenTime.Load() != utcNow.Add(11*time.Second).UnixMicro() {
			t.Fail()
		}
	})

	t.Run("should descrease free tokens on each call", func(t *testing.T) {
		rateLimiter := NewLockFreeRateLimiter(1*time.Second, 2, DefaultTimeProvider)
		rateLimiter.Try()
		if rateLimiter.freeTokens.Load() != 1 {
			t.Fail()
		}
		rateLimiter.Try()
		if rateLimiter.freeTokens.Load() != 0 {
			t.Fail()
		}
	})

	t.Run("shoul permit execution when next token generation has arrived", func(t *testing.T) {
		tokenPerUnit := 10 * time.Millisecond
		rateLimiter := NewLockFreeRateLimiter(tokenPerUnit, 0, DefaultTimeProvider)
		time.Sleep(tokenPerUnit)
		ok, _ := rateLimiter.Try()

		if !ok {
			t.Fail()
		}
	})

	t.Run("should permit execution as many times as there are free tokens", func(t *testing.T) {
		rateLimiter := NewLockFreeRateLimiter(1*time.Hour, 2, DefaultTimeProvider)
		ok1, _ := rateLimiter.Try()
		ok2, _ := rateLimiter.Try()
		ok3, _ := rateLimiter.Try()

		if !ok1 || !ok2 || ok3 {
			t.Fail()
		}
	})
}

func exhaustFreeTokens(rl lockFreeRateLimiter) {
	for {
		if rl.freeTokens.Load() <= 0 {
			break
		}
		rl.Try()
	}
}
