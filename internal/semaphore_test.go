package internal

import (
	"testing"
)

func TestConcurrencysem(t *testing.T) {
	t.Run("sem should remain in a consistent state", func(t *testing.T) {
		sem := NewSemaphore(3)
		sem.Wait()
		sem.Wait()
		if !sem.TryWait() {
			t.Fail()
		}
		if sem.TryWait() {
			t.Fail()
		}

		sem.Release()
		sem.Release()
		sem.Release()

		if !sem.TryWait() {
			t.Fail()
		}
		sem.Release()
	})

	t.Run("attempt to pass a sem should fail when there are no free slots", func(t *testing.T) {
		sem := NewSemaphore(1)
		sem.Wait()
		defer sem.Release()

		if sem.TryWait() {
			defer sem.Release()
			t.Fail()
		}
	})

	t.Run("should limit the level of concurrency to a single unit", func(t *testing.T) {
		sem := NewSemaphore(1)
		flag := false
		sem.Wait()

		go func() {
			defer sem.Release()
			flag = true
		}()

		sem.Wait()
		sem.Release()

		if !flag {
			t.Fail()
		}
	})

	t.Run("`Release` call should affect available slots", func(t *testing.T) {
		sem := NewSemaphore(0)
		sem.Release()
		sem.Release()

		if sem.threshold != 2 {
			t.Fail()
		}
	})

	t.Run("should decrement/increment counter of available slots", func(t *testing.T) {
		sem := NewSemaphore(2)
		sem.Wait()
		slots := sem.threshold
		sem.Release()

		if slots != 1 || sem.threshold != 2 {
			t.Fail()
		}
	})
}
