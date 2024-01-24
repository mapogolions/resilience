package internal

import (
	"sync"
	"time"
)

type circuitState[T any] interface {
	IsCircuitOpen() bool
	Success()
	Failure(T, error)
}

type circuitStateOpen[T any] struct {
	cb *circuitBreaker[T]
}

func (s *circuitStateOpen[T]) IsCircuitOpen() bool {
	return true
}

func (s *circuitStateOpen[T]) Success() {
}

func (s *circuitStateOpen[T]) Failure(T, error) {
}

type circuitStateClosed[T any] struct {
	cb *circuitBreaker[T]
}

func (s *circuitStateClosed[T]) IsCircuitOpen() bool {
	return s.cb.timeProvider.UtcNow().Before(s.cb.brokenTill)
}

func (s *circuitStateClosed[T]) Success() {
}

func (s *circuitStateClosed[T]) Failure(result T, err error) {
	s.cb.lastResult = result
	s.cb.lastErr = err
	s.cb.consecutiveFailures++
	if s.cb.consecutiveFailures >= s.cb.breakAfter {
		s.cb.state = s.cb.open
		s.cb.brokenTill = s.cb.timeProvider.UtcNow().Add(s.cb.breakDuration)
	}
}

type circuitStateHalfOpen[T any] struct {
	cb *circuitBreaker[T]
}

func (s *circuitStateHalfOpen[T]) IsCircuitOpen() bool {
	return false
}

func (s *circuitStateHalfOpen[T]) Success() {
	s.cb.state = s.cb.closed
	s.cb.reset()
}

func (s *circuitStateHalfOpen[T]) Failure(result T, err error) {
	s.cb.lastResult = result
	s.cb.lastErr = err
	s.cb.state = s.cb.closed
	s.cb.brokenTill = s.cb.timeProvider.UtcNow().Add(s.cb.breakDuration)
}

type circuitBreaker[T any] struct {
	sync.Mutex
	open                circuitState[T]
	closed              circuitState[T]
	halfOpen            circuitState[T]
	state               circuitState[T]
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
	cb := &circuitBreaker[T]{
		breakAfter:    breakAfter,
		breakDuration: breakDuration,
		timeProvider:  timeProvider}
	cb.open = &circuitStateOpen[T]{cb}
	cb.halfOpen = &circuitStateHalfOpen[T]{cb}
	cb.closed = &circuitStateClosed[T]{cb}
	cb.state = cb.closed
	return cb
}

func (cb *circuitBreaker[T]) IsCircuitOpen() bool {
	if cb.state != cb.open {
		return false
	}
	cb.Lock()
	defer cb.Unlock()
	return cb.state.IsCircuitOpen()
}

func (cb *circuitBreaker[T]) Success() {
	cb.Lock()
	defer cb.Unlock()
	cb.state.Success()
}

func (cb *circuitBreaker[T]) Failure(result T, err error) {
	cb.Lock()
	defer cb.Unlock()
	cb.state.Failure(result, err)
}
