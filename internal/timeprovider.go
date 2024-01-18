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
