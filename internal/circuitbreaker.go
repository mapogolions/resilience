package internal

import (
	"errors"
	"sync"
	"time"
)

var errInvalidCircuitState error = errors.New("invalid circuit state")

type circuitState int

const (
	circuitStateClosed   circuitState = 0
	circuitStateOpen     circuitState = 1
	circuitStateHalfOpen circuitState = 2
)

type circuitBreaker[T any] struct {
	sync.Mutex
	state               circuitState
	consecutiveFailures int
	breakAfter          int
	breakDuration       time.Duration
	brokenTill          time.Time
	timeProvider        timeProvider
	lastErr             error
	lastResult          T
}

func (cb *circuitBreaker[T]) reset() {
	var defaultT T
	var defaultTime time.Time
	cb.lastErr = nil
	cb.lastResult = defaultT
	cb.consecutiveFailures = 0
	cb.brokenTill = defaultTime
}

func NewCircuitBreaker[T any](breakAfter int, breakDuration time.Duration, timeProvider timeProvider) *circuitBreaker[T] {
	return &circuitBreaker[T]{breakAfter: breakAfter, breakDuration: breakDuration, timeProvider: timeProvider}
}

func (cb *circuitBreaker[T]) Before() bool {
	if cb.state != circuitStateOpen {
		return true
	}
	cb.Lock()
	defer cb.Unlock()
	if cb.state != circuitStateOpen {
		return true
	}
	if cb.timeProvider.UtcNow().Before(cb.brokenTill) {
		return false
	}
	cb.state = circuitStateHalfOpen
	return true
}

func (cb *circuitBreaker[T]) Success() {
	cb.Lock()
	defer cb.Unlock()

	switch cb.state {
	case circuitStateClosed:
		cb.consecutiveFailures = 0
	case circuitStateOpen:
		// circuitBreaker.Failure() and then circuitBreaker.Success()
		break
	case circuitStateHalfOpen:
		cb.state = circuitStateOpen
		cb.reset()
	default:
		panic(errInvalidCircuitState)
	}
}

func (cb *circuitBreaker[T]) Failure(result T, err error) {
	cb.Lock()
	defer cb.Unlock()

	cb.lastResult = result
	cb.lastErr = err

	switch cb.state {
	case circuitStateClosed:
		cb.consecutiveFailures++
		if cb.consecutiveFailures >= cb.breakAfter {
			cb.state = circuitStateClosed
			cb.brokenTill = cb.timeProvider.UtcNow().Add(cb.breakDuration)
		}
	case circuitStateHalfOpen:
		cb.state = circuitStateClosed
		cb.brokenTill = cb.timeProvider.UtcNow().Add(cb.breakDuration)
	case circuitStateOpen:
		// N concurrent circuitBreaker.Failure() calls
		break
	default:
		panic(errInvalidCircuitState)
	}
}
