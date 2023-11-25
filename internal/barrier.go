package internal

import (
	"sync"
)

type Barrier struct {
	cond  *sync.Cond
	limit int
}

func NewBarrier(limit int) *Barrier {
	if limit <= 0 {
		panic("limit should greater than zero")
	}
	mutex := sync.Mutex{}
	return &Barrier{cond: sync.NewCond(&mutex), limit: limit}
}

func (b *Barrier) TryWait() bool {
	b.cond.L.Lock()
	if b.limit > 0 {
		b.limit--
		b.cond.L.Unlock()
		return true
	}
	b.cond.L.Unlock()
	return false
}

func (b *Barrier) Wait() {
	b.cond.L.Lock()
	if b.limit > 0 {
		b.limit--
		b.cond.L.Unlock()
		return
	}
	for {
		if b.limit > 0 {
			break
		}
		b.cond.Wait()
	}
	b.limit--
	b.cond.L.Unlock()
}

func (b *Barrier) Release() {
	b.cond.L.Lock()
	if b.limit > 0 {
		b.limit--
		b.cond.Broadcast()
	}
	b.cond.L.Unlock()
}
