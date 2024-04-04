package policy

import (
	"context"
	"strings"
	"testing"
)

func TestPolicy(t *testing.T) {
	f := func(_ context.Context, s string, count int) (string, error) {
		return strings.Repeat(s, count), nil
	}

	t.Run("should be able to apply policy to n-argumnets function using adapter", func(t *testing.T) {
		policy := NewIdentityPolicy[tuple2[string, int], string]()
		r, err := policy(context.Background(), adapter(f), tuple2[string, int]{item0: ".", item1: 3})

		if err != nil {
			t.Fail()
		}
		if r != "..." {
			t.Fail()
		}
	})

	t.Run("should be able to apply policy to n-arguments function using function expression", func(t *testing.T) {
		policy := NewIdentityPolicy[interface{}, string]()
		r, err := policy(context.Background(), func(ctx context.Context, _ interface{}) (string, error) {
			return f(ctx, ".", 3)
		}, nil)

		if err != nil {
			t.Fail()
		}
		if r != "..." {
			t.Fail()
		}
	})
}

type tuple2[A any, B any] struct {
	item0 A
	item1 B
}

func adapter[A any, B any, T any](f func(context.Context, A, B) (T, error)) func(context.Context, tuple2[A, B]) (T, error) {
	return func(ctx context.Context, p tuple2[A, B]) (T, error) {
		return f(ctx, p.item0, p.item1)
	}
}
