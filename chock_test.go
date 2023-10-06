package chock_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/kitd/chock"
)

// Run this file with `go test -test.v ./...` to see sample error output

func TestCherr(t *testing.T) {
	err_msg := "An error occured"
	ctx_msg := "running TestCherr "

	old_err := fmt.Errorf(err_msg)
	new_err := Wrap(old_err)
	new_err.Context(ctx_msg)
	message := new_err.Error()
	if !strings.Contains(message, err_msg) {
		t.Errorf("Error did not contain expected string '%s'", err_msg)
	} else if !strings.Contains(message, ctx_msg) {
		t.Errorf("Error did not contain expected string '%s'", ctx_msg)
	} else {
		t.Logf("%v\n", new_err)
	}
}

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
		t.Logf("%v\n", r.Context("running TestFailureWithContext").Unwrap())
	}
}

func TestFlags(t *testing.T) {
	defer func() {
		IncludeContext = true
		IncludeSource = false
	}()

	IncludeContext = false
	IncludeSource = true
	if r := myOtherFunctionThatFails(); !r.Failed() {
		t.Errorf("result succeeded. It should have failed")
	} else {
		err := r.Context("running TestFlags").Unwrap()
		errStr := err.Error()
		if strings.Contains(errStr, "Context:") {
			t.Errorf("error contains context. It should not have")
		}
		if !strings.Contains(errStr, "- =>") {
			t.Errorf("error does not contain source. It should have")
		}
		t.Logf("%v\n", err)
	}
}

func myFunctionThatFails[T any]() Result[T] {
	return Failure[T](fmt.Errorf("An error has occurred"))
}

func myOtherFunctionThatFails() Result[int] {
	r := myFunctionThatFails[int]()
	if r.Failed() {
		return r.Context("calling myOtherFunctionThatFails")
	} else {
		return r
	}
}
