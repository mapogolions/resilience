package internal

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("should close circuit whe it is half open and next call succeeded", func(t *testing.T) {
		// Arrange
		breakAfter := 2
		breakDuration := 2 * time.Second
		timeProvider := NewFakeTimeProvider()
		cb := NewCircuitBreaker[int](breakAfter, breakDuration, timeProvider)

		// Act
		cb.Failure(-1, errors.New("err1"))
		cb.Failure(-2, errors.New("err2"))
		timeProvider.Advance(breakDuration)
		cb.IsCircuitOpen() // half open state
		cb.Success()

		// Assert
		if cb.IsCircuitOpen() || cb.consecutiveFailures != 0 || cb.state != circuitStateClosed {
			t.Fail()
		}
	})

	t.Run("should open circuit when it is half open and failure happens", func(t *testing.T) {
		// Arrange
		breakAfter := 2
		breakDuration := 2 * time.Second
		timeProvider := NewFakeTimeProvider()
		cb := NewCircuitBreaker[int](breakAfter, breakDuration, timeProvider)

		// Act
		cb.Failure(-1, errors.New("err1"))
		cb.Failure(-2, errors.New("err2"))
		timeProvider.Advance(breakDuration)
		cb.IsCircuitOpen() // half open state
		cb.Failure(-3, errors.New("err3"))

		// Assert
		if !cb.IsCircuitOpen() {
			t.Fail()
		}
	})

	t.Run("should half open circuit after specified break period", func(t *testing.T) {
		breakAfter := 1
		breakDuration := 2 * time.Second
		timeProvider := NewFakeTimeProvider()
		cb := NewCircuitBreaker[int](breakAfter, breakDuration, timeProvider)
		cb.Failure(-1, errors.New("err1"))
		timeProvider.Advance(breakDuration)

		if cb.IsCircuitOpen() || cb.state != circuitStateHalfOpen {
			t.Fail()
		}
	})

	t.Run("circuit should stay open until specified time", func(t *testing.T) {
		breakAfter := 2
		breakDuration := 2 * time.Second
		timeProvider := NewFakeTimeProvider()
		cb := NewCircuitBreaker[int](breakAfter, breakDuration, timeProvider)
		cb.Failure(-1, errors.New("error1"))
		cb.Failure(-2, errors.New("error2"))
		timeProvider.Advance(1 * time.Second)

		if !cb.IsCircuitOpen() {
			t.Fail()
		}
	})

	t.Run("should open circuit after specified consecutive failures", func(t *testing.T) {
		breakAfter := 2
		breakDuration := 2 * time.Second
		cb := NewCircuitBreaker[int](breakAfter, breakDuration, DefaultTimeProvider)
		cb.Failure(-1, errors.New("error1"))
		cb.Failure(-2, errors.New("error2"))

		if !cb.IsCircuitOpen() {
			t.Fail()
		}
	})

	t.Run("circuit should be closed at the begining", func(t *testing.T) {
		breakAfter := 2
		breakDuration := 2 * time.Second
		cb := NewCircuitBreaker[int](breakAfter, breakDuration, DefaultTimeProvider)

		if cb.IsCircuitOpen() {
			t.Fail()
		}
	})
}
