package internal

import (
	"sync"
)

type Barrier struct {
	cond  *sync.Cond
	limit int
	slots int
}

func NewBarrier(limit int) *Barrier {
	if limit <= 0 {
		panic("limit should greater than zero")
	}
	mutex := sync.Mutex{}
	return &Barrier{cond: sync.NewCond(&mutex), limit: limit, slots: limit}
}

func (b *Barrier) TryWait() bool {
	b.cond.L.Lock()
	if b.slots > 0 {
		b.slots--
		b.cond.L.Unlock()
		return true
	}
	b.cond.L.Unlock()
	return false
}

func (b *Barrier) Wait() {
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

func (b *Barrier) Release() {
	b.cond.L.Lock()
	if b.slots < b.limit {
		b.slots++
		b.cond.Broadcast()
	}
	b.cond.L.Unlock()
}
