package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDebounceFirst(t *testing.T) {
	t.Run("concurrent calls should not block while callback is running", func(t *testing.T) {
		t.Parallel()

		// arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		policy := NewDebounceFirstPolicy[string, int](10 * time.Second)

		done := make(chan any)
		firstCall := make(chan any)

		go func() {
			defer close(done)

			// Signal once the first invocation starts executing `f`.
			policy(ctx, func(ctx context.Context, s string) (int, error) {
				close(firstCall)
				return spin[string, int](ctx, s)
			}, "foo")
		}()

		<-firstCall

		// act
		result, err := policy(ctx, spin, "foo")

		// assert
		if !errors.Is(err, ErrDebounced) || result != 0 {
			t.Fail()
		}

		cancel()
		<-done
	})

	t.Run("should debounce second call within window when context already cancelled", func(t *testing.T) {
		t.Parallel()

		// arrange
		policy := NewDebounceFirstPolicy[int, int](5 * time.Second)
		f := newSliceIndexer([]int{10, 20, 30})
		ctx, cancel := context.WithCancel(context.Background())

		// act
		cancel()
		result, err := policy(ctx, f, 0)
		result, err = policy(ctx, f, 0)

		// assert
		if !errors.Is(err, ErrDebounced) || result != 0 {
			t.Fail()
		}
	})

	t.Run("should return cancelled error when context already cancelled", func(t *testing.T) {
		// arrange
		policy := NewDebounceFirstPolicy[int, int](5 * time.Second)
		f := newSliceIndexer([]int{10, 20, 30})
		ctx, cancel := context.WithCancel(context.Background())

		// act
		cancel()
		result, err := policy(ctx, f, 0)

		// assert
		if !errors.Is(err, context.Canceled) || result != 0 {
			t.Fail()
		}
	})

	t.Run("should not suppress calls after debounce window", func(t *testing.T) {
		// arrange
		d := 200 * time.Millisecond
		timer := time.NewTimer(d * 2)
		policy := NewDebounceFirstPolicy[int, int](d)
		f := newSliceIndexer([]int{10, 20, 30})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// act + assert
		result, err := policy(ctx, f, 1)
		if err != nil || result != 20 {
			t.Fail()
		}

		<-timer.C

		result, err = policy(ctx, f, 1)
		if err != nil || result != 20 {
			t.Fail()
		}
	})

	t.Run("should suppress calls within debounce window", func(t *testing.T) {
		// arrange
		policy := NewDebounceFirstPolicy[int, int](5 * time.Second)
		f := newSliceIndexer([]int{10, 20, 30})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// act + assert
		result, err := policy(ctx, f, 1)

		if err != nil || result != 20 {
			t.Fail()
		}

		result, err = policy(ctx, f, 1000)
		if !errors.Is(err, ErrDebounced) {
			t.Fail()
		}
	})
}

func spin[S, T any](ctx context.Context, _ S) (T, error) {
	var zero T
	for {
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func newSliceIndexer[T any](items []T) func(context.Context, int) (T, error) {
	return func(ctx context.Context, i int) (T, error) {
		var zero T
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		if i < 0 || i > len(items) {
			return zero, errors.New("out of range exception")
		}
		if items == nil {
			return zero, errors.New("null pointer exception")
		}
		return items[i], nil
	}
}
