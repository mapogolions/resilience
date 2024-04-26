### Bulkhead policy

Bulkhead policy involves limiting the number of units of concurrency (threads, goroutines) that can access a particular resource concurrently.

The policy allows you to set 2 types of resource access thresholds:

- CONCURRENCY - sets the number of units of concurrency that can simultaneously access the resource.
- QUEUE - the number of units of concurrency that will get a chance to access the resource after some waiting.
