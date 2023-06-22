# Chock

A Rust-like `Result[T]` handling package, that wraps either a value of type T, or an error.

Actual arrors are wrapped in an internal error that incorporates a stack trace, and allows context to be added before the result is returned, eg:

```
    chock_test.go:33: An error has occurred
        Context:
        - calling myOtherFunctionThatFails
        - running TestFailureWithContext
        Stack:
        - (/home/kit/dev/chock/chock_test.go:37) github.com/kitd/chock_test.myFunctionThatFails[...]
        - (/home/kit/dev/chock/chock_test.go:41) github.com/kitd/chock_test.myOtherFunctionThatFails
        - (/home/kit/dev/chock/chock_test.go:29) github.com/kitd/chock_test.TestFailureWithContext
        - (/usr/local/go/src/testing/testing.go:1576) testing.tRunner
```

An MIT license is applied.