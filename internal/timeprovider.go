package internal

import "time"

type timeProvider interface {
	UtcNow() time.Time
}

type timeProviderFunc func() time.Time

func (f timeProviderFunc) UtcNow() time.Time {
	return f()
}

var DefaultTimeProvider timeProviderFunc = func() time.Time {
	return time.Now().UTC()
}

type fakeTimeProvider struct {
	now time.Time
}

func NewFakeTimeProvider() *fakeTimeProvider {
	return &fakeTimeProvider{now: time.Now().UTC()}
}

func (tp *fakeTimeProvider) UtcNow() time.Time {
	return tp.now
}

func (tp *fakeTimeProvider) Advance(delta time.Duration) time.Time {
	prev := tp.now
	tp.now = tp.now.Add(delta)
	return prev
}
