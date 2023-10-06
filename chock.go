package chock

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var (
	IncludeContext = true
	IncludeStack   = true
	IncludeSource  = false
)

const (
	ENV_INCL_CTX    = "CHOCK_INCL_CTX"
	ENV_INCL_STACK  = "CHOCK_INCL_STACK"
	ENV_INCL_SOURCE = "CHOCK_INCL_SOURCE"
)

func init() {
	loadFlagFromEnv(&IncludeContext, ENV_INCL_CTX)
	loadFlagFromEnv(&IncludeStack, ENV_INCL_STACK)
	loadFlagFromEnv(&IncludeSource, ENV_INCL_SOURCE)
}

func loadFlagFromEnv(flag *bool, envVar string) {
	if b, e := strconv.ParseBool(os.Getenv(envVar)); e == nil {
		*flag = b
	}
}

type ErrorWithContext interface {
	error
	Context(ctx string)
}

type cherr struct {
	cause   error
	stack   []string
	context []string
	source  []string
}

func (n *cherr) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\nCause: \"%s\"\n", n.cause.Error()))
	if IncludeContext && len(n.context) > 0 {
		sb.WriteString("Context:\n")
		for _, ctx := range n.context {
			sb.WriteString(fmt.Sprintf("- %s\n", ctx))
		}
	}
	if IncludeStack {
		sb.WriteString("Stack:\n")
		for _, frame := range n.stack {
			sb.WriteString(fmt.Sprintf("- %s\n", frame))
		}
	}
	if IncludeSource {
		sb.WriteString("Source:\n")
		for _, line := range n.source {
			sb.WriteString(fmt.Sprintf("- %s\n", line))
		}
	}
	return sb.String()
}

func (n *cherr) Unwrap() error {
	return n.cause
}

func (n *cherr) Context(ctx string) {
	n.context = append(n.context, ctx)
}

func Wrap(cause error) ErrorWithContext {
	err := &cherr{
		cause: cause,
	}
	var ptrs [64]uintptr
	count := runtime.Callers(2, ptrs[:])
	if count > 0 {
		frames := runtime.CallersFrames(ptrs[:count])
		top := true
		for frame, more := frames.Next(); more; frame, more = frames.Next() {
			err.stack = append(err.stack, fmt.Sprintf("(%s:%d) %s", frame.File, frame.Line, frame.Function))
			if top {
				if IncludeSource {
					if file, e := os.Open(frame.File); e == nil {
						defer file.Close()
						scanner := bufio.NewScanner(file)
						scanner.Split(bufio.ScanLines)
						for lineNo := 1; lineNo < frame.Line-1; lineNo++ {
							scanner.Scan()
						}
						scanner.Scan()
						err.source = append(err.source, "   "+scanner.Text())
						scanner.Scan()
						err.source = append(err.source, "=> "+scanner.Text())
						if scanner.Scan() {
							err.source = append(err.source, "   "+scanner.Text())
						}
					}
				}
				top = false
			}
		}
	}

	return err
}

type Result[T any] interface {
	error
	Failed() bool
	Value() T
	Context(ctx string) Result[T]
	Unwrap() error
}

type resultImpl[T any] struct {
	value   T
	failure ErrorWithContext
}

func (r *resultImpl[T]) Failed() bool {
	return r.failure != nil
}

func (r *resultImpl[T]) Value() T {
	return r.value
}

func (r *resultImpl[T]) Context(ctx string) Result[T] {
	r.failure.Context(ctx)
	return r
}

func (r *resultImpl[T]) Error() string {
	if r.Failed() {
		return r.failure.Error()
	} else {
		return ""
	}
}

func (r *resultImpl[T]) Unwrap() error {
	return r.failure
}

func Success[T any](value T) Result[T] {
	return &resultImpl[T]{value, nil}
}

func Failure[T any](cause error) Result[T] {
	return &resultImpl[T]{*new(T), Wrap(cause)}
}
