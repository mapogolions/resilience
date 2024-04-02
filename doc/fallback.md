### Fallback

The fallback policy executes the initial function and, before finalizing the result, transfers it to a designated fallback function. This secondary function then evaluates how to manage the outcome, deciding whether to retain it unchanged or intervene and provide a more suitable alternative.


How to provide custom fallback function.

```golang
policy.NewFallbackPolicy[S, T](func(ctx context.Context, o policy.Outcome[T]) (T, error) {
    // determines how to handle the given result - whether to return it as it is or to intercept it and return a more appropriate value
})
```

The simplest form of fallback, which returns the result as it is, is `IdentityFallback`.

```golang
policy.NewFallbackPolicy[S, T](policy.IdentityFallback[T])
```

The library provides another method called `NewPanicFallbackPolicy`. It allows intercepting errors associated with calling the `panic` method.

```golang
policy.NewPanicFallbackPolicy[S, T](/* your fallback */)
```
