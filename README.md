# Chock

A `Result[T]` handling package for Go, that wraps either a value of type `T`, or an error.

Typical usage:

```go
import "github.com/kitd/chock"

func someFunctionThatMightFail(arg0 string) chock.Result[int] {
    if intVal, err := somepkg.MyIntFunction(arg0); err != nil {
        return chock.Failure[int](err).Context("arg0", arg0)
    } else {
        return chock.Success(intVal)
    }
}

func anotherFunction() chock.Result[int] {
    if r := someFunctionThatMightFail("xyz"); r.Failed() {
        return r.Context("foo", "bar")
    } else {
        doSomethingWith(r.Value())
    }
}
```

For functions that return a tuple of `(type, error)`, you can call `chock.ResultOf(...)`. Eg:
```go
if myFile := chock.ResultOf(ioutils.ReadFile("myFile")); myFile.Failed() {
    return myFile.Context("Trying to read myFile")
} else {
    fmt.println(string(myFile.Value()[:]))
}
```

Actual errors are wrapped in an internal error that incorporates a stack trace (from the point where `chock.Failure(cause)` is called), and allows context to be added before the result is returned, eg:
```
    chock_test.go:33: 
        Cause: "An error has occurred"
        Context:
        {
            "foo": "bar",
            "arg0": "xyz"
        }
        Stack:
        - (/home/kit/dev/chock/chock.go:103) github.com/kitd/chock.Failure[...]
        - (/home/kit/dev/chock/chock_test.go:37) github.com/kitd/chock_test.myFunctionThatFails[...]
        - (/home/kit/dev/chock/chock_test.go:41) github.com/kitd/chock_test.myOtherFunctionThatFails
        - (/home/kit/dev/chock/chock_test.go:29) github.com/kitd/chock_test.TestFailureWithContext
        - (/usr/local/go/src/testing/testing.go:1576) testing.tRunner
```
The file name and line number are formatted to make them clickable in VSCode, allowing you to open the source file at the error line in a single click.  

You can switch off the display of the stack by setting the `CHOCK_INCL_STACK` env var to `false`. Similarly, the display of the context info can be controlled using the `CHOCK_INCL_CTX` env var. 

If you set the `CHOCK_INCL_SOURCE` env var to true, it will display the source line of the top stack frame, along with the preceding and succeeding lines. Eg:
```
    Source:
    -    func Failure[T any](cause error) Result[T] {
    - =>    return &resultImpl[T]{*new(T), Wrap(cause)}
    -    }
```
Note that this only really makes sense in testing as the source code will probably not be available in production.

An MIT license is applied.