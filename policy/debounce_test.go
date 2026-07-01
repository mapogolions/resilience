package policy

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDebounceLast(t *testing.T) {
	t.Run("should return value after specified timeout", func(t *testing.T) {
		// arrange
		d := 100 * time.Millisecond
		timer := time.NewTimer(d)
		policy := NewDebouncePolicy[int, int](d, DebounceLast)
		f := newSliceIndexer([]int{10, 20, 30})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// act + assert
		result, err := policy(ctx, f, 1)

		if !errors.Is(err, ErrDebounced) || result != 0 {
			t.Fail()
		}

		<-timer.C

		result, err = policy(ctx, f, 1)
		if err != nil {
			t.Fail()
		}
	})

	t.Run("should return debounced error on the first call", func(t *testing.T) {
		// arrange
		policy := NewDebouncePolicy[int, int](50*time.Millisecond, DebounceLast)
		f := newSliceIndexer([]int{10, 20, 30})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// act
		result, err := policy(ctx, f, 1)

		// assert
		if !errors.Is(err, ErrDebounced) || result != 0 {
			t.Fail()
		}
	})
}

func TestDebounceFirst(t *testing.T) {
	t.Run("should debounce second call within window when context already cancelled", func(t *testing.T) {
		// arrange
		policy := NewDebouncePolicy[int, int](5*time.Second, DebounceFirst)
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
		policy := NewDebouncePolicy[int, int](5*time.Second, DebounceFirst)
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
		d := 500 * time.Millisecond
		timer := time.NewTimer(d)
		policy := NewDebouncePolicy[int, int](d, DebounceFirst)
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
		policy := NewDebouncePolicy[int, int](5*time.Second, DebounceFirst)
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
