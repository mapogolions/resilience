package resilience

import (
	"context"
)

type Policy[S any, T any] func(context.Context, func(context.Context, S) (T, error), S) (T, error)
