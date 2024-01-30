package internal

import (
	"errors"
	"sync"
	"time"
)

// Do not use state pattern because the code  becomes less readable
type circuitState int

const (
	circuitStateClosed   = 0
	circuitStateOpen     = 1
	circuitStateHalfOpen = 2
)

var ErrInvalidCircuitState = errors.New("invalid circuit state")

type circuitBreaker[T any] struct {
	sync.Mutex
	state               circuitState
	consecutiveFailures int
	breakAfter          int
	breakDuration       time.Duration
	breakTill           time.Time
	timeProvider        timeProvider
	lastResult          T
	lastErr             error
}

func NewCircuitBreaker[T any](breakAfter int, breakDuration time.Duration, timeProvider timeProvider) *circuitBreaker[T] {
	return &circuitBreaker[T]{
		breakAfter:    breakAfter,
		breakDuration: breakDuration,
		timeProvider:  timeProvider,
	}
}

func (cb *circuitBreaker[T]) LastResult() T {
	return cb.lastResult
}

func (cb *circuitBreaker[T]) LastError() error {
	return cb.lastErr
}

func (cb *circuitBreaker[T]) SetBreakTill() {
	cb.breakTill = cb.timeProvider.UtcNow().Add(cb.breakDuration)
}

func (cb *circuitBreaker[T]) IsCircuitOpen() bool {
	if cb.state != circuitStateOpen {
		return false
	}
	cb.Lock()
	defer cb.Unlock()
	if cb.state != circuitStateOpen {
		return false
	}
	if cb.timeProvider.UtcNow().Before(cb.breakTill) {
		return true
	}
	cb.state = circuitStateHalfOpen
	return false
}

func (cb *circuitBreaker[T]) Success() {
	cb.Lock()
	defer cb.Unlock()
	switch cb.state {
	case circuitStateClosed:
		cb.consecutiveFailures = 0
	case circuitStateHalfOpen:
		cb.state = circuitStateClosed
		cb.consecutiveFailures = 0
		cb.breakTill = time.Time{}
	case circuitStateOpen:
		break
	default:
		panic(ErrInvalidCircuitState)
	}
}

func (cb *circuitBreaker[T]) Failure(result T, err error) {
	cb.Lock()
	defer cb.Unlock()
	cb.lastResult = result
	cb.lastErr = err
	switch cb.state {
	case circuitStateOpen:
		break
	case circuitStateHalfOpen:
		cb.state = circuitStateOpen
		cb.SetBreakTill()
	case circuitStateClosed:
		cb.consecutiveFailures++
		if cb.consecutiveFailures >= cb.breakAfter {
			cb.state = circuitStateOpen
			cb.SetBreakTill()
		}
	default:
		panic(ErrInvalidCircuitState)
	}
}
