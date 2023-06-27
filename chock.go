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

func (n *cherr) addContext(ctx string) {
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

type Result[T any] interface {
	error
	Failed() bool
	Value() T
	Context(ctx string) Result[T]
	Unwrap() error
}

type resultImpl[T any] struct {
	value   T
	failure error
}

func (r *resultImpl[T]) Failed() bool {
	return r.failure != nil
}

func (r *resultImpl[T]) Value() T {
	return r.value
}

func (r *resultImpl[T]) Context(ctx string) Result[T] {
	r.failure.(*cherr).addContext(ctx)
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
