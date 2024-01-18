package internal

import "time"

type TimeProvider interface {
	UtcNow() time.Time
}

type TimeProviderFunc func() time.Time

func (f TimeProviderFunc) UtcNow() time.Time {
	return f()
}

var DefaultTimeProvider TimeProviderFunc = func() time.Time {
	return time.Now().UTC()
}

type fakeTimeProvider struct {
	now time.Time
}

func (tp *fakeTimeProvider) UtcNow() time.Time {
	return tp.now
}

func (tp *fakeTimeProvider) Advance(delta time.Duration) time.Time {
	prev := tp.now
	tp.now = tp.now.Add(delta)
	return prev
}
