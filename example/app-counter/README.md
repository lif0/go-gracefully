# Example

In this example, we increment a counter and write its value to the file `counter.txt` whenever it is divisible by 10 without a remainder. Below is an illustration of how this application behaves with and without the `go-gracefully` library.

## Without go-gracefully

```log
% go run .
Last counter: 0
Press any key to increment (press 'q' to quit)
counter: 1
counter: 2
counter: 3
counter: 4
counter: 5
counter: 6
counter: 7
counter: 8
counter: 9
counter: 10
counter: 11
counter: 12
counter: 13
counter: 14
^Csignal: interrupt
```

However, if we check the contents of the `counter.txt` file, we will see `10`. This happens because we did not have time to call `flush()` before termination.

## With go-gracefully

We must implement the following method for the counter object:
```
GracefulShutdown(ctx context.Context) error
```

Now, register the object using `go-gracefully.Register(counter)` and observe:

```log
go run .
Last counter: 0
Press any key to increment (press 'q' to quit)
counter: 1
counter: 2
counter: 3
counter: 4
counter: 5
counter: 6
counter: 7
counter: 8
counter: 9
counter: 10
counter: 11
counter: 12
counter: 13
^C2025/10/11 20:42:58 gogracefully: Received system signal - interrupt
2025/10/11 20:42:58 gogracefully: Graceful shutdown completed. Use gogracefully.GlobalErrors for checks errors
App finish
```

Now, if we check the contents of the `counter.txt` file, we will see `13`. This way, we have preserved the intermediate state and avoided losing it.