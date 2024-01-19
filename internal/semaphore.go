package internal

import (
	"sync"
)

type semaphore struct {
	cond      *sync.Cond
	threshold int
}

func NewSemaphore(threshold int) *semaphore {
	if threshold < 0 {
		panic("semaphore initial value must be >= 0")
	}
	mutex := sync.Mutex{}
	return &semaphore{cond: sync.NewCond(&mutex), threshold: threshold}
}

func (b *semaphore) TryWait() bool {
	b.cond.L.Lock()
	if b.threshold > 0 {
		b.threshold--
		b.cond.L.Unlock()
		return true
	}
	b.cond.L.Unlock()
	return false
}

func (b *semaphore) Wait() {
	b.cond.L.Lock()
	if b.threshold > 0 {
		b.threshold--
		b.cond.L.Unlock()
		return
	}

	for {
		b.cond.Wait()
		if b.threshold > 0 {
			break
		}
	}
	b.threshold--
	b.cond.L.Unlock()
}

func (b *semaphore) Release() {
	b.cond.L.Lock()
	v := b.threshold
	b.threshold++
	if v == 0 { // or b.threshold == 1
		b.cond.Broadcast()
	}
	b.cond.L.Unlock()
}
