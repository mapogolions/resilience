package bulkhead

import (
	"context"
	"sync"
	"testing"

	"github.com/mapogolions/resilience/internal"
)

func TestBulkhead(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		bulkhead := NewBulkhead[int, int](2, 3)
		barrier := make(chan struct{})
		f := func(ctx context.Context, index int) (int, error) {
			<-barrier
			return index, nil
		}
		allDone := sync.WaitGroup{}
		allDone.Add(10)
		runningCounter := internal.NewCounter()
		errorCounter := internal.NewCounter()
		for i := 0; i < 10; i++ {
			go func(i int) {
				defer allDone.Done()
				if _, cur := runningCounter.Increment(); cur == 10 {
					close(barrier)
				}
				_, err := bulkhead(context.Background(), f, i)
				if err != nil {
					errorCounter.Increment()
				}
			}(i)
		}
		allDone.Wait()

		if errorCounter.Value() != 5 {
			t.Fail()
		}
	})
}
