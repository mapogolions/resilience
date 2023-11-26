package internal

import "sync"

type Counter struct {
	mutex *sync.Mutex
	value int
}

func NewCounter() *Counter {
	return &Counter{mutex: &sync.Mutex{}}
}

func (c *Counter) Value() int {
	return c.value
}

func (c *Counter) Increment() (int, int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	prev := c.value
	c.value++
	cur := c.value
	return prev, cur
}

func (c *Counter) Decrement() (int, int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	prev := c.value
	c.value--
	cur := c.value
	return prev, cur
}
