# Chock

A Rust-like `Result[T]` handling package for Go, that wraps either a value of type `T`, or an error.

Typical usage:

```go
import "github.com/kitd/chock"

func someFunctionThatMightFail(arg0 string) chock.Result[int] {
    if intVal, err := external.IntFunction(arg0); err != nil {
        return chock.Failure[int](err).Context(fmt.Sprintf("arg0=%s", arg0))
    } else {
        return chock.Success(intVal)
    }
}

func anotherFunction() chock.Result[int] {
    if r := someFunctionThatMightFail("xyz"); r.Failed() {
        return r.Context("calling anotherFunction")
    } else {
        doSomethingWith(r.Value())
    }
}
```

Actual errors are wrapped in an internal error that incorporates a stack trace (from the point where `chock.Wrap(cause)` is called), and allows context to be added before the result is returned, eg:
```
    chock_test.go:33: An error has occurred
        Context:
        - calling myOtherFunctionThatFails
        - running TestFailureWithContext
        Stack:
        - (/home/kit/dev/chock/chock.go:103) github.com/kitd/chock.Failure[...]
        - (/home/kit/dev/chock/chock_test.go:37) github.com/kitd/chock_test.myFunctionThatFails[...]
        - (/home/kit/dev/chock/chock_test.go:41) github.com/kitd/chock_test.myOtherFunctionThatFails
        - (/home/kit/dev/chock/chock_test.go:29) github.com/kitd/chock_test.TestFailureWithContext
        - (/usr/local/go/src/testing/testing.go:1576) testing.tRunner
```

If you only want the enhanced error features without using the `Result[T]` type, you can call the `chock.Wrap(cause)` function. You can add context to the resulting error, which will appear with a stack trace when output via the `error.Error()` function.

An MIT license is applied.