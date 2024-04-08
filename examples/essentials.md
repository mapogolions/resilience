
### Essentials

Policy is a function that executes another function provided as an input argument based on encapsulated internal logic. It extends or modifies the behavior of the original function without directly altering its code."

```golang
func[S any, T any](context.Context, func(context.Context, S) (T, error), S) (T, error)
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
r, err := policy(context.Background(), f, x)
```

If you need to apply a policy to a function that takes no arguments or more than one argument,

```golang
func f(context.Context, A, B) (T, error) {
	panic("not implemented")
}
```
you can use the following trick

```golang
p := policy.NewRetryPolicy[interface{}, T](policy.NewRetryCountOnErrorCondition[T](2))
r, err := p(context.Background(), func(ctx context.Context, _ interface{}) (T, error) {
	return f(ctx, A, B)
}, nil)
```

or write an [adapter](../policy/policy_test.go).
