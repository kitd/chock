package chock

import (
	"fmt"
	"runtime"
	"strings"
)

type cherr struct {
	cause   error
	stack   []string
	context []string
}

func (n *cherr) Error() string {
	var sb strings.Builder
	sb.WriteString(n.cause.Error() + "\n")
	if len(n.context) > 0 {
		sb.WriteString("Context:\n")
		for _, ctx := range n.context {
			sb.WriteString("- " + ctx + "\n")
		}
	}
	sb.WriteString("Stack:\n")
	for _, frame := range n.stack {
		sb.WriteString("- " + frame + "\n")
	}
	return sb.String()
}

func (n *cherr) Unwrap() error {
	return n.cause
}

func (n *cherr) AddContext(ctx string) {
	n.context = append(n.context, ctx)
}

func Wrap(cause error) error {
	err := &cherr{
		cause: cause,
	}
	pcs := make([]uintptr, 64)
	count := runtime.Callers(3, pcs)
	if count > 0 {
		frames := runtime.CallersFrames(pcs[:count])
		for frame, more := frames.Next(); more; frame, more = frames.Next() {
			err.stack = append(err.stack, fmt.Sprintf("(%s:%d) %s", frame.File, frame.Line, frame.Function))
		}
	}

	return err
}

type Result[T any] struct {
	Value   T
	failure error
}

func (r *Result[T]) Failed() bool {
	return r.failure != nil
}

func (r *Result[T]) With(ctx string) *Result[T] {
	r.failure.(*cherr).AddContext(ctx)
	return r
}

func (r *Result[T]) Unwrap() error {
	return r.failure
}

func Success[T any](value T) *Result[T] {
	return &Result[T]{value, nil}
}

func Failure[T any](cause error) *Result[T] {
	return &Result[T]{*new(T), Wrap(cause)}
}
