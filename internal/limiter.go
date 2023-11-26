package internal

import (
	"sync"
)

// acts like a semaphore
type ConcurrencyLimiter struct {
	cond        *sync.Cond
	concurrency int
	slots       int
}

func NewConcurrencyLimiter(concurrency int) *ConcurrencyLimiter {
	if concurrency <= 0 {
		panic("concurrency level should be greater than zero")
	}
	mutex := sync.Mutex{}
	return &ConcurrencyLimiter{cond: sync.NewCond(&mutex), concurrency: concurrency, slots: concurrency}
}

func (b *ConcurrencyLimiter) TryWait() bool {
	b.cond.L.Lock()
	if b.slots > 0 {
		b.slots--
		b.cond.L.Unlock()
		return true
	}
	b.cond.L.Unlock()
	return false
}

func (b *ConcurrencyLimiter) Wait() {
	b.cond.L.Lock()
	if b.slots > 0 {
		b.slots--
		b.cond.L.Unlock()
		return
	}
	for {
		b.cond.Wait()
		if b.slots > 0 {
			break
		}
	}
	b.slots--
	b.cond.L.Unlock()
}

func (b *ConcurrencyLimiter) Release() {
	b.cond.L.Lock()
	if b.slots < b.concurrency {
		b.slots++
		b.cond.Broadcast()
	}
	b.cond.L.Unlock()
}
