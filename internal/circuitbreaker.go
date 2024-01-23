package internal

import (
	"errors"
	"sync"
	"time"
)

var errInvalidCircuiteState error = errors.New("invalid circuite state")

type curcuiteState int

const (
	curcuiteStateClosed   curcuiteState = 0
	curcuiteStateOpen     curcuiteState = 1
	curcuiteStateHalfOpen curcuiteState = 2
)

type curcuiteBreaker[T any] struct {
	sync.Mutex
	state               curcuiteState
	consecutiveFailures int
	breakAfter          int
	breakDuration       time.Duration
	brokenTill          time.Time
	timeProvider        timeProvider
	lastErr             error
	lastResult          T
}

func (cb *curcuiteBreaker[T]) reset() {
	var defaultT T
	var defaultTime time.Time
	cb.lastErr = nil
	cb.lastResult = defaultT
	cb.consecutiveFailures = 0
	cb.brokenTill = defaultTime
}

func NewCircuiteBreaker[T any](breakAfter int, breakDuration time.Duration, timeProvider timeProvider) *curcuiteBreaker[T] {
	return &curcuiteBreaker[T]{breakAfter: breakAfter, breakDuration: breakDuration, timeProvider: timeProvider}
}

func (cb *curcuiteBreaker[T]) Before() bool {
	if cb.state != curcuiteStateOpen {
		return true
	}
	cb.Lock()
	defer cb.Unlock()
	if cb.state != curcuiteStateOpen {
		return true
	}
	if cb.timeProvider.UtcNow().Before(cb.brokenTill) {
		return false
	}
	cb.state = curcuiteStateHalfOpen
	return true
}

func (cb *curcuiteBreaker[T]) Success() {
	cb.Lock()
	defer cb.Unlock()

	switch cb.state {
	case curcuiteStateClosed:
		cb.consecutiveFailures = 0
	case curcuiteStateOpen:
		// circuiteBreaker.Failure() and then circuiteBreaker.Success()
		break
	case curcuiteStateHalfOpen:
		cb.state = curcuiteStateOpen
		cb.reset()
	default:
		panic(errInvalidCircuiteState)
	}
}

func (cb *curcuiteBreaker[T]) Failure(result T, err error) {
	cb.Lock()
	defer cb.Unlock()

	cb.lastResult = result
	cb.lastErr = err

	switch cb.state {
	case curcuiteStateClosed:
		cb.consecutiveFailures++
		if cb.consecutiveFailures >= cb.breakAfter {
			cb.state = curcuiteStateClosed
			cb.brokenTill = cb.timeProvider.UtcNow().Add(cb.breakDuration)
		}
	case curcuiteStateHalfOpen:
		cb.state = curcuiteStateClosed
		cb.brokenTill = cb.timeProvider.UtcNow().Add(cb.breakDuration)
	case curcuiteStateOpen:
		// N concurrent circuiteBreaker.Failure() calls
		break
	default:
		panic(errInvalidCircuiteState)
	}
}
