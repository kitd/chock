package chock

import (
	"bufio"
	"encoding/json"
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
	if val, err := strconv.ParseBool(os.Getenv(envVar)); err == nil {
		*flag = val
	}
}

type ErrorWithContext interface {
	error
	Context(key string, value any) ErrorWithContext
	Unwrap() error
}

type chockError struct {
	cause   error
	stack   []string
	context map[string]any
	source  []string
}

func (n *chockError) Error() string {
	sb := &strings.Builder{}
	sb.WriteString("\nCause: \"")
	sb.WriteString(n.cause.Error())
	sb.WriteString("\"\n")

	if IncludeContext && len(n.context) > 0 {
		json, _ := json.MarshalIndent(n.context, "", "  ")
		sb.WriteString("Context:\n")
		sb.WriteString(fmt.Sprint(string(json)))
		sb.WriteString("\n")
	}
	if IncludeStack {
		writeStrings(sb, "Stack", n.stack)
	}
	if IncludeSource {
		writeStrings(sb, "Source", n.source)
	}
	return sb.String()
}

func writeStrings(sb *strings.Builder, label string, lines []string) {
	sb.WriteString(label + ":\n")
	for _, line := range lines {
		sb.WriteString("- ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}
}

func (n *chockError) Unwrap() error {
	return n.cause
}

func (n *chockError) Context(key string, value any) ErrorWithContext {
	n.context[key] = value
	return n
}

func Wrap(cause error) ErrorWithContext {
	err := &chockError{
		cause:   cause,
		context: map[string]any{},
	}
	var ptrs [64]uintptr
	count := runtime.Callers(2, ptrs[:]) // '2' skips frames until the caller of 'Wrap'
	if count > 0 {
		frames := runtime.CallersFrames(ptrs[:count])
		top := true
		for frame, hasMore := frames.Next(); hasMore; frame, hasMore = frames.Next() {
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
	Failed() bool
	Failure() ErrorWithContext
	Context(key string, value any) Result[T]
	Value() T
}

type resultImpl[T any] struct {
	value   T
	failure ErrorWithContext
}

func (r *resultImpl[T]) Failed() bool {
	return r.failure != nil
}

func (r *resultImpl[T]) Failure() ErrorWithContext {
	return r.failure
}

func (r *resultImpl[T]) Value() T {
	return r.value
}

func (r *resultImpl[T]) Context(key string, value any) Result[T] {
	r.failure.Context(key, value)
	return r
}

func (r *resultImpl[T]) Error() string {
	if r.Failed() {
		return r.failure.Error()
	} else {
		return ""
	}
}

func Success[T any](value T) Result[T] {
	return &resultImpl[T]{value, nil}
}

func Failure[T any](cause error) Result[T] {
	return &resultImpl[T]{*new(T), Wrap(cause)}
}
