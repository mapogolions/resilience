package internal

import (
	"sync/atomic"
)

type Counter struct {
	value atomic.Int64
}

func NewCounter() *Counter {
	return &Counter{value: atomic.Int64{}}
}

func (c *Counter) Value() int64 {
	return c.value.Load()
}

func (c *Counter) Increment() int64 {
	cur := c.value.Add(1)
	return cur
}

func (c *Counter) Decrement() int64 {
	cur := c.value.Add(-1)
	return cur
}
