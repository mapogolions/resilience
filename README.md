### Resilience

Inspired by [The Polly Project](https://www.thepollyproject.org)

The motivation behind the project is learning by creating. As I delved into the technical details of the Polly library, it seemed to me that the same ideas could be expressed a little more simply.

#### Policy essentials

Policy is a function decorator that take another function as input and return a new function that extends or modifies the behavior of the original function without directly altering its code.

The most generic form of function decorator

```golang
func[S any, T any](func(s S) (T, error), S) (T, error)
```

The above definition is almost identical to the one used inside this library. The library forces the client to use the `context.Context` parameter.

```golang
func[S any, T any](context.Context, func(context.Conte, S) (T, error), s S) (T, error)
```
You are free to implement your own policy. For example, in the snippet below, the `delay` policy executes the original function with a specified delay. This functionality could be particularly useful during unit tests.

```golang
func delay[S any, T any](d time.Duration) resilience.Policy[S, T] {
	return func(ctx context.Context, f func(context.Context, S) (T, error), s S) (T, error) {
		time.Sleep(d)
		return f(ctx, s)
	}
}
```

The library offers primitives for combining policies to create complex behavior.

```golang
policy := policy.Compose[string, []byte](g, f) // compose(g, f) = x => g(f(x))
	policy.NewRetryPolicy[string, []byte](policy.NewRetryCountOnErrorCondition[[]byte](3)), // g (outer)
	policy.NewTimeoutPolicy[string, []byte](4*time.Second, policy.OptimisticTimeoutPolicy), // f (inner)
)
r, err := policy(context.Background(), request, "...")
```

Or you could use `Pipeline` to use `andThen` behaviour.

```golang
policy := policy.Pipeline[string, []byte](g, f) // pipeline(f, g) = x => g(f(x))
	policy.NewTimeoutPolicy[string, []byte](4*time.Second, policy.OptimisticTimeoutPolicy), // f (inner)
	policy.NewRetryPolicy[string, []byte](policy.NewRetryCountOnErrorCondition[[]byte](3)), // g (outer)
)
r, err := policy(context.Background(), request, "...")
```
