package policy

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mapogolions/resilience/internal"
)

func TestBulkhead(t *testing.T) {
	t.Run("should limit concurrency level", func(t *testing.T) {
		testCases := []struct {
			concurrency    int
			queue          int
			total          int
			expectedErrors int
		}{
			{concurrency: 1, queue: 4, total: 10, expectedErrors: 5},
			{concurrency: 4, queue: 1, total: 10, expectedErrors: 5},
			{concurrency: 1, queue: 4, total: 5, expectedErrors: 0},
			{concurrency: 4, queue: 1, total: 5, expectedErrors: 0},
		}
		for _, testCase := range testCases {
			policy := NewBulkheadPolicy[int, int](testCase.concurrency, testCase.queue)
			barrier := make(chan struct{})
			f := func(ctx context.Context, index int) (int, error) {
				<-barrier
				return index, nil
			}
			allDone := sync.WaitGroup{}
			allDone.Add(testCase.total)
			runningCounter := internal.NewCounter()
			errorCounter := internal.NewCounter()
			for i := 0; i < testCase.total; i++ {
				go func(i int) {
					defer allDone.Done()
					if cur := runningCounter.Increment(); cur == int64(testCase.total) {
						time.AfterFunc(100*time.Millisecond, func() { close(barrier) })
					}
					_, err := policy.Apply(context.Background(), f, i)
					if err != nil {
						errorCounter.Increment()
					}
				}(i)
			}
			allDone.Wait()

			if errorCounter.Value() != int64(testCase.expectedErrors) {
				t.Fail()
			}
		}
	})
}
