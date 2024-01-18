package internal

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
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
