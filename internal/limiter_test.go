package internal

import (
	"testing"
)

func TestConcurrencyLimiter(t *testing.T) {
	t.Run("limiter should remain in a consistent state", func(t *testing.T) {
		limiter := NewConcurrencyLimiter(3)
		limiter.Wait()
		limiter.Wait()
		if !limiter.TryWait() {
			t.Fail()
		}
		if limiter.TryWait() {
			t.Fail()
		}

		limiter.Release()
		limiter.Release()
		limiter.Release()

		if !limiter.TryWait() {
			t.Fail()
		}
		limiter.Release()
	})

	t.Run("attempt to pass a limiter should fail when there are no free slots", func(t *testing.T) {
		limiter := NewConcurrencyLimiter(1)
		limiter.Wait()
		defer limiter.Release()

		if limiter.TryWait() {
			defer limiter.Release()
			t.Fail()
		}
	})

	t.Run("should limit the level of concurrency to a single unit", func(t *testing.T) {
		limiter := NewConcurrencyLimiter(1)
		flag := false
		limiter.Wait()

		go func() {
			defer limiter.Release()
			flag = true
		}()

		limiter.Wait()
		limiter.Release()

		if !flag {
			t.Fail()
		}
	})

	t.Run("`Release` call should not affect available slots", func(t *testing.T) {
		limiter := NewConcurrencyLimiter(2)
		limiter.Release()

		if limiter.slots != 2 {
			t.Fail()
		}
	})

	t.Run("should decrement/increment counter of available slots", func(t *testing.T) {
		limiter := NewConcurrencyLimiter(2)
		limiter.Wait()
		slots := limiter.slots
		limiter.Release()

		if slots != 1 || limiter.slots != 2 {
			t.Fail()
		}
	})
}
