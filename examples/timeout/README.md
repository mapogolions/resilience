### Timeout policy

Timeout policy limits the execution of a given function in time. Suppose we have a function `f` that accepts context.Context as one of its parameters.

```golang
func f(context.Context, S) (T, error) {}
```

The standard way to limit execution of a given function in time.

```golang
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
f(ctx, s)
```

#### Pessimistic timeout policy

What's wrong with this approach? This approach is optimistic. The function 'f', from the example above, may ignore context, rarely checking or not
checking it at all regarding the deadline. Pessimistic timeout policy can help you solve the mentioned problem.

```golang
p := policy.NewTimeoutPolicy[S, T](timeout, policy.PessimisticTimeoutPolicy)
r, err := p(context.Background(), f, s)
```

Please note that this policy doesn't magically change the code of the `f` function, making it consider the context deadline. The function moves the computation of the `f` function into a separate goroutine. If the goroutine allocated for the computation does not complete within the specified time interval, then the waiting concurrency unit(goroutine) continues execution. It should be noted that the allocated goroutine does not disappear until the 'f' function finishes execution.


#### Optimistic timeout policy

Optimistic timeout policy follows the standard approach but with a few unique characteristics.

```golang
p := policy.NewTimeoutPolicy[S, T](timeout, policy.OptimisticTimeoutPolicy)
r, err := p(context.Background(), f, s)
```
