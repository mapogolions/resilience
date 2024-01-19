package internal

import (
	"sync/atomic"
)

type counter struct {
	value atomic.Int64
}

func NewCounter() *counter {
	return &counter{value: atomic.Int64{}}
}

func (c *counter) Value() int64 {
	return c.value.Load()
}

func (c *counter) Increment() int64 {
	cur := c.value.Add(1)
	return cur
}

func (c *counter) Decrement() int64 {
	cur := c.value.Add(-1)
	return cur
}
