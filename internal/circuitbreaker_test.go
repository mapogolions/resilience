package internal

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("should close circuit when it is half open and next call succeeded", func(t *testing.T) {
		// Arrange
		breakDuration := 2 * time.Second
		timeProvider := NewFakeTimeProvider()
		cb := NewCircuitBreaker[int](1, breakDuration, timeProvider)

		// Act
		cb.Failure(-1, errors.New("err1"))
		timeProvider.Advance(breakDuration)
		cb.IsCircuitOpen() // half open state
		cb.Success()

		// Assert
		if cb.IsCircuitOpen() {
			t.Fail()
		}

		if cb.consecutiveFailures != 0 || cb.state != circuitStateClosed {
			t.Fail()
		}

		if (cb.breakTill != time.Time{}) {
			t.Fail()
		}

	})

	t.Run("should open circuit when it is half open and next call failed", func(t *testing.T) {
		// Arrange
		breakDuration := 2 * time.Second
		timeProvider := NewFakeTimeProvider()
		cb := NewCircuitBreaker[int](1, breakDuration, timeProvider)

		// Act
		cb.Failure(-1, errors.New("err1"))
		timeProvider.Advance(breakDuration)
		cb.IsCircuitOpen() // half open state
		cb.Failure(-2, errors.New("err2"))

		// Assert
		if !cb.IsCircuitOpen() {
			t.Fail()
		}
		if timeProvider.UtcNow().Add(breakDuration) != cb.breakTill {
			t.Fail()
		}
	})

	t.Run("should half open circuit after specified break period", func(t *testing.T) {
		breakDuration := 2 * time.Second
		timeProvider := NewFakeTimeProvider()
		cb := NewCircuitBreaker[int](1, breakDuration, timeProvider)
		cb.Failure(-1, errors.New("err1"))
		timeProvider.Advance(breakDuration)

		if cb.IsCircuitOpen() || cb.state != circuitStateHalfOpen {
			t.Fail()
		}
	})

	t.Run("circuit should stay open until specified time", func(t *testing.T) {
		timeProvider := NewFakeTimeProvider()
		cb := NewCircuitBreaker[int](2, 2*time.Second, timeProvider)
		cb.Failure(-1, errors.New("err1"))
		cb.Failure(-2, errors.New("err2"))
		timeProvider.Advance(1 * time.Second)

		if !cb.IsCircuitOpen() {
			t.Fail()
		}
	})

	t.Run("should stop increasing consecutive failures when failure threshold reached", func(t *testing.T) {
		cb := NewCircuitBreaker[int](2, 2*time.Second, DefaultTimeProvider)
		cb.Failure(-1, errors.New("err1"))
		cb.Failure(-2, errors.New("err2"))
		cb.Failure(-3, errors.New("err3"))

		if cb.consecutiveFailures != 2 {
			t.Fail()
		}
	})

	t.Run("should open circuit after specified consecutive failures", func(t *testing.T) {
		cb := NewCircuitBreaker[int](1, 2*time.Second, DefaultTimeProvider)
		cb.Failure(-1, errors.New("err1"))
		cb.Failure(-2, errors.New("err2"))

		if !cb.IsCircuitOpen() {
			t.Fail()
		}
	})

	t.Run("should be closed when consecutive failures is not enought", func(t *testing.T) {
		cb := NewCircuitBreaker[int](2, 1*time.Second, DefaultTimeProvider)
		cb.Failure(-1, errors.New("err1"))

		if cb.IsCircuitOpen() || cb.consecutiveFailures != 1 {
			t.Fail()
		}
	})

	t.Run("circuit should be closed by default", func(t *testing.T) {
		cb := NewCircuitBreaker[int](2, 2*time.Second, DefaultTimeProvider)

		if cb.IsCircuitOpen() {
			t.Fail()
		}
	})
}
