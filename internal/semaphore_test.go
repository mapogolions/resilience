package internal

import (
	"context"
	"strings"
	"testing"
)

func TestConcurrencysem(t *testing.T) {
	t.Run("sem should remain in a consistent state", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sem := NewBoundedSemaphore(3)

		sem.Wait(ctx)
		sem.Wait(ctx)
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

	t.Run("should be able to cancel wait", func(t *testing.T) {
		sem := NewBoundedSemaphore(0)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			cancel()
		}()
		sem.Wait(ctx)
	})

	t.Run("attempt to pass a sem should fail when there are no free slots", func(t *testing.T) {
		sem := NewBoundedSemaphore(1)
		sem.Wait(context.Background())
		defer sem.Release()

		if sem.TryWait() {
			defer sem.Release()
			t.Fail()
		}
	})

	t.Run("Release without matching Wait should panic with expected message", func(t *testing.T) {
		sem := NewBoundedSemaphore(0)

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic, got none")
			}
			msg, ok := r.(string)
			if !ok || !strings.Contains(msg, "release without matching wait") {
				t.Fatalf("unexpected panic value: %v", r)
			}
		}()

		sem.Release()
	})
}
