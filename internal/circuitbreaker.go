package internal

import (
	"errors"
	"sync"
	"time"
)

// Do not use state pattern because the code  becomes less readable
type circuitState int

const (
	circuitStateClosed   circuitState = 0
	circuitStateOpen     circuitState = 1
	circuitStateHalfOpen circuitState = 2
)

var ErrInvalidCircuitState = errors.New("invalid circuit state")

type circuitBreaker[T any] struct {
	sync.Mutex
	state               circuitState
	consecutiveFailures int
	failureThreshold    int
	breakDuration       time.Duration
	breakTill           time.Time
	timeProvider        timeProvider
	lastResult          T
	lastErr             error
}

func NewCircuitBreaker[T any](
	failureThreshold int,
	breakDuration time.Duration,
	timeProvider timeProvider) *circuitBreaker[T] {

	return &circuitBreaker[T]{
		state:            circuitStateClosed,
		failureThreshold: failureThreshold,
		breakDuration:    breakDuration,
		timeProvider:     timeProvider,
	}
}

func (cb *circuitBreaker[T]) setBreakTill() {
	cb.breakTill = cb.timeProvider.UtcNow().Add(cb.breakDuration)
}

func (cb *circuitBreaker[T]) TryAcquire() bool {
	cb.Lock()
	defer cb.Unlock()

	if cb.state == circuitStateOpen {
		if cb.timeProvider.UtcNow().Before(cb.breakTill) {
			return false
		}
		cb.state = circuitStateHalfOpen
	}
	return true
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
		cb.setBreakTill()
	case circuitStateClosed:
		cb.consecutiveFailures++
		if cb.consecutiveFailures >= cb.failureThreshold {
			cb.state = circuitStateOpen
			cb.setBreakTill()
		}
	default:
		panic(ErrInvalidCircuitState)
	}
}
