package chock_test

import (
	"fmt"
	"testing"

	. "github.com/kitd/chock"
)

// Run this file with `go test -test.v ./...` to see sample error output

func TestSuccess(t *testing.T) {
	r := Success(42)

	if r.Failed() || r.Value() != 42 {
		t.Errorf("result failed. It should have passed with 42")
	}
}

func TestPlainFailure(t *testing.T) {
	if r := myFunctionThatFails[int](); !r.Failed() {
		t.Errorf("result succeeded. It should have failed")
	} else {
		t.Logf("%v\n", r.Unwrap())
	}
}

func TestFailureWithContext(t *testing.T) {
	if r := myOtherFunctionThatFails(); !r.Failed() {
		t.Errorf("result succeeded. It should have failed")
	} else {
		t.Logf("%v\n", r.With("running TestFailureWithContext").Unwrap())
	}
}

func myFunctionThatFails[T any]() Result[T] {
	return Failure[T](fmt.Errorf("An error has occurred"))
}

func myOtherFunctionThatFails() Result[int] {
	r := myFunctionThatFails[int]()
	if r.Failed() {
		return r.With("calling myOtherFunctionThatFails")
	} else {
		return r
	}
}
