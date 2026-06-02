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
	for envVar := range TraceFlags {
		if val, err := strconv.ParseBool(os.Getenv(envVar)); err == nil {
			TraceFlags[envVar] = val
		}
	}
}

func writeStrings(sb *strings.Builder, label string, lines []string, block bool) {
	if block {
		sb.WriteString(label + ": |\n")
	} else {
		sb.WriteString(label + ":\n")
	}
	for _, line := range lines {
		if block {
			sb.WriteString("  ")
		} else {
			sb.WriteString("- ")
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
}

type Result[T any] interface {
	Failed() bool
}

type Success[T any] struct {
	Value T
}

func Ok[T any](value T) Result[T] {
	return &Success[T]{value}
}

func (r *Success[T]) Failed() bool {
	return false
}

type Failure struct {
	cause   error
	context []string
	stack   []string
	source  []string
}

func (r *Failure) Failed() bool {
	return true
}

func (r *Failure) Unwrap() error {
	return r.cause
}

func (r *Failure) WithContext(val string) *Failure {
	r.context = append(r.context, val)
	return r
}

func (r *Failure) WithContextf(format string, args ...any) *Failure {
	return r.WithContext(fmt.Sprintf(format, args...))
}

func (r *Failure) Error() string {
	sb := &strings.Builder{}
	sb.WriteString("\nCause: \"")
	sb.WriteString(r.cause.Error())
	sb.WriteString("\"\n")

	if TraceFlags[ENV_INCL_CTX] && len(r.context) > 0 {
		writeStrings(sb, "Context", r.context, false)
	}
	if TraceFlags[ENV_INCL_STACK] {
		writeStrings(sb, "Stacktrace", r.stack, false)
	}
	if TraceFlags[ENV_INCL_SOURCE] {
		writeStrings(sb, "Source", r.source, false)
	}
	return sb.String()
}

func Wrap(cause error) *Failure {

	if failure, ok := cause.(*Failure); ok {
		return failure
	}

	err := &Failure{cause, nil, nil, nil}
	if TraceFlags[ENV_INCL_STACK] {
		var ptrs [64]uintptr
		count := runtime.Callers(2, ptrs[:]) // '2' skips frames until our caller
		if count > 0 {
			frames := runtime.CallersFrames(ptrs[:count])
			top := true
			for frame, hasMore := frames.Next(); hasMore; frame, hasMore = frames.Next() {
				err.stack = append(err.stack, fmt.Sprintf("(%s:%d) %s", frame.File, frame.Line, frame.Function))
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
							err.source = append(err.source, "  "+scanner.Text())
							scanner.Scan()
							err.source = append(err.source, ">>"+scanner.Text())
							if scanner.Scan() {
								err.source = append(err.source, "  "+scanner.Text())
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

func ResultOf[T any](value T, err error) Result[T] {
	if err != nil {
		return Wrap(err)
	} else {
		return Ok(value)
	}
}

func Failed[T any](result Result[T]) *Failure {
	if result.Failed() {
		return result.(*Failure)
	} else {
		return nil
	}
}

func Succeeded[T any](result Result[T]) *Success[T] {
	if !result.Failed() {
		return result.(*Success[T])
	} else {
		return nil
	}
}
