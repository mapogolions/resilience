
### Essentials

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

The library offers primitives such as `Compose` and `Pipeline` for combining policies to create complex behavior.

```golang
policy := policy.Compose[S, T]( // compose(g, f) = x => g(f(x))
	policy.NewRetryPolicy[S, T](policy.NewRetryCountOnErrorCondition[T](3)), // g (outer)
	policy.NewTimeoutPolicy[S, T](4*time.Second, policy.OptimisticTimeoutPolicy), // f (inner)
)
r, err := policy(context.Background(), request, "...")
```
