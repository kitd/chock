package chock_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/kitd/chock"
)

// Run this file with `go test -test.v ./...` to see sample error output

func TestSuccess(t *testing.T) {
	r := Ok(42)

	if r.Failed() || r.(*Success[int]).Value != 42 {
		t.Errorf("result failed. It should have passed with 42")
	}
}

func TestPlainFailure(t *testing.T) {
	if r := sourceOfFailure(); !r.Failed() {
		t.Errorf("result succeeded. It should have failed")
	} else {
		t.Logf("%v\n", r.(*Failure).Error())
	}
}

func TestFailureWithContext(t *testing.T) {
	if r := intermediateFunc(); !r.Failed() {
		t.Errorf("result succeeded. It should have failed")
	} else {
		t.Logf("%v\n", r.(*Failure).Context("running TestFailureWithContext").Error())
	}
}

func TestFlags(t *testing.T) {
	defer func() {
		TraceFlags[ENV_INCL_CTX] = true
		TraceFlags[ENV_INCL_SOURCE] = false
	}()

	TraceFlags[ENV_INCL_CTX] = false
	TraceFlags[ENV_INCL_SOURCE] = true
	if r := intermediateFunc(); !r.Failed() {
		t.Errorf("result succeeded. It should have failed")
	} else {
		err := r.(*Failure).Context("running TestFlags")
		errStr := err.Error()
		if strings.Contains(errStr, "Context:") {
			t.Errorf("error contains context. It should not have")
		}
		if !strings.Contains(errStr, "Source: |") {
			t.Errorf("error does not contain source. It should have")
		}
		t.Logf("%v\n", err)
	}
}

func TestResultOf(t *testing.T) {
	if result := ResultOf(funcThatSucceeds()); result.Failed() {
		t.Error("Test funcThatSucceeds should have passed")
	} else {
		t.Logf("Succeeded as expected: %d", result.(*Success[int]).Value)
	}

	if result := ResultOf(funcThatFails()); !result.Failed() {
		t.Error("Test funcThatFails should have failed")
	} else {
		t.Logf("Failed as expected: %v", result.(*Failure).Context("Doing funcThatFails"))
	}
}

func sourceOfFailure() Result {
	return Error(fmt.Errorf("An error has occurred"))
}

func intermediateFunc() Result {
	r := sourceOfFailure()
	if r.Failed() {
		return r.(*Failure).Context("calling myOtherFunctionThatFails")
	} else {
		return r
	}
}

func funcThatSucceeds() (int, error) {
	return 1, nil
}

func funcThatFails() (int, error) {
	return 0, fmt.Errorf("Test Error")
}
