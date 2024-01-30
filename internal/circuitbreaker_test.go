package internal

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("circuit should be half open after specified break period", func(t *testing.T) {
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
