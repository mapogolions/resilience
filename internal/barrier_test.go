package internal

import (
	"testing"
)

func TestBarrier(t *testing.T) {
	t.Run("an attempt to pass the barrier should fail if there are no free slots", func(t *testing.T) {
		barrier := NewBarrier(1)
		flag := true
		barrier.Wait()

		go func() {
			defer barrier.Release()
			flag = barrier.TryWait()
		}()

		barrier.Wait()
		barrier.Release()

		if flag {
			t.Fail()
		}
	})

	t.Run("should limit the level of concurrency to a single unit", func(t *testing.T) {
		barrier := NewBarrier(1)
		flag := false
		barrier.Wait()

		go func() {
			defer barrier.Release()
			flag = true
		}()

		barrier.Wait()
		barrier.Release()

		if !flag {
			t.Fail()
		}
	})

	t.Run("`Release` call should not affect available slots", func(t *testing.T) {
		barrier := NewBarrier(2)
		barrier.Release()

		if barrier.slots != 2 {
			t.Fail()
		}
	})

	t.Run("should decrement/increment counter of available slots", func(t *testing.T) {
		barrier := NewBarrier(2)
		barrier.Wait()
		slots := barrier.slots
		barrier.Release()

		if slots != 1 || barrier.slots != 2 {
			t.Fail()
		}
	})
}
