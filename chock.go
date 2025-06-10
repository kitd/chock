package chock

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	ENV_INCL_CTX    = "CHOCK_INCL_CTX"
	ENV_INCL_STACK  = "CHOCK_INCL_STACK"
	ENV_INCL_SOURCE = "CHOCK_INCL_SOURCE"
)

var (
	TraceFlags map[string]bool = map[string]bool{
		ENV_INCL_CTX:    true,
		ENV_INCL_STACK:  true,
		ENV_INCL_SOURCE: false,
	}
)

func init() {
	RefreshConfig()
}

func RefreshConfig() {
	for envVar, _ := range TraceFlags {
		if val, err := strconv.ParseBool(os.Getenv(envVar)); err == nil {
			TraceFlags[envVar] = val
		}
	}
}

func writeStrings(sb *strings.Builder, label string, lines []string) {
	sb.WriteString(label + ":\n")
	for _, line := range lines {
		sb.WriteString("- ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}
}

type chockError struct {
	err    error
	stack  []string
	source []string
}

type Result[T any] struct {
	value   T
	failure *chockError
	context []string
}

func (r *Result[T]) Failed() bool {
	return r.failure != nil
}

func (r *Result[T]) Value() T {
	return r.value
}

func (r *Result[T]) Unwrap() error {
	if r.Failed() {
		return r.failure.err
	} else {
		return nil
	}
}

func (r *Result[T]) Context(val string) *Result[T] {
	r.context = append(r.context, val)
	return r
}

func (r *Result[T]) Contextf(format string, args ...any) *Result[T] {
	r.context = append(r.context, fmt.Sprintf(format, args...))
	return r
}

func (r *Result[T]) Error() string {
	if r.Failed() {
		sb := &strings.Builder{}
		sb.WriteString("\nCause: \"")
		sb.WriteString(r.Unwrap().Error())
		sb.WriteString("\"\n")

		if TraceFlags[ENV_INCL_CTX] && len(r.context) > 0 {
			writeStrings(sb, "Context", r.context)
		}
		if TraceFlags[ENV_INCL_STACK] {
			writeStrings(sb, "Stack", r.failure.stack)
		}
		if TraceFlags[ENV_INCL_SOURCE] {
			writeStrings(sb, "Source", r.failure.source)
		}
		return sb.String()
	} else {
		return ""
	}
}

func Success[T any](value T) *Result[T] {
	return &Result[T]{value, nil, nil}
}

func Failure[T any](cause error) *Result[T] {
	var zero T
	err := &Result[T]{zero, &chockError{cause, nil, nil}, nil}
	if TraceFlags[ENV_INCL_STACK] {
		var ptrs [64]uintptr
		count := runtime.Callers(2, ptrs[:]) // '2' skips frames until our caller
		if count > 0 {
			frames := runtime.CallersFrames(ptrs[:count])
			top := true
			for frame, hasMore := frames.Next(); hasMore; frame, hasMore = frames.Next() {
				err.failure.stack = append(err.failure.stack, fmt.Sprintf("(%s:%d) %s", frame.File, frame.Line, frame.Function))
				if top {
					if TraceFlags[ENV_INCL_SOURCE] {
						if file, e := os.Open(frame.File); e == nil {
							defer file.Close()
							scanner := bufio.NewScanner(file)
							scanner.Split(bufio.ScanLines)
							for lineNo := 1; lineNo < frame.Line-1; lineNo++ {
								scanner.Scan()
							}
							scanner.Scan()
							err.failure.source = append(err.failure.source, "   "+scanner.Text())
							scanner.Scan()
							err.failure.source = append(err.failure.source, "=> "+scanner.Text())
							if scanner.Scan() {
								err.failure.source = append(err.failure.source, "   "+scanner.Text())
							}
						}
					}
					top = false
				}
			}
		}
	}
	return err
}

func ResultOf[T any](value T, err error) *Result[T] {
	if err != nil {
		return Failure[T](err)
	} else {
		return Success(value)
	}
}
