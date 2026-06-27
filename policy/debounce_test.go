package policy

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDebounceFirst(t *testing.T) {
	t.Run("should suppress calls within debounce window", func(t *testing.T) {
		// arrange
		policy := NewDebouncePolicy[int, int](1*time.Second, DebounceFirst)
		circuite := newSliceIndexer([]int{10, 20, 30})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// act + assert
		result, err := policy(ctx, circuite, 1)

		if err != nil || result != 20 {
			t.Fail()
		}

		result, err = policy(ctx, circuite, 1000)
		if !errors.Is(err, ErrDebounced) {
			t.Fail()
		}
	})
}

func newSliceIndexer[T any](items []T) func(context.Context, int) (T, error) {
	return func(_ context.Context, i int) (T, error) {
		var v T
		if i < 0 || i > len(items) {
			return v, errors.New("out of range exception")
		}
		if items == nil {
			return v, errors.New("null pointer exception")
		}
		return items[i], nil
	}
}
