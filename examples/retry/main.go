package main

import (
	"fmt"
	"time"

	"github.com/mapogolions/resilience/policy"
)

func main() {
	{
		policy.NewRetryPolicy[string, []byte](
			policy.NewRetryCountOnErrorCondition[[]byte](10),
		)
	}

	{
		policy.NewRetryPolicy[string, []byte](
			policy.NewRetryCountOnErrorWithDelayCondition[[]byte](
				3,
				func(i int) time.Duration {
					return time.Duration(i) * time.Second
				},
			),
		)
	}
	fmt.Println("foo")
}
