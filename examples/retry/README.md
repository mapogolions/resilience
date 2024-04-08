### Retry policy

The retry policy executes a given function at least once. After the initial execution, the policy decides whether to repeat the execution of the function or return a result.
The retry policy assesses whether further retries are required based on a specified retry condition.

The library provides convenient helper functions for creating common retry conditions:

- retry a specified number of times (`N`) in case of an error occurrence.

```golang
policy.NewRetryPolicy[T](policy.NewRetryCountOnErrorCondition[T](N))
```

- retry a specified number (`n`) of times with delays that adapt based on the current attempt

```golang
policy.NewRetryPolicy[S, T](
    policy.NewRetryCountOnErrorWithDelayCondition[T](
        N,
        func(i int) time.Duration {
            return time.Duration(i) * time.Second
        },
    ),
)
```
