package assert

import (
	"fmt"
	"testing"
)

type fakeTestingT struct {
	failNowed bool
	failed    bool
	msgs      []string
}

func (f *fakeTestingT) FailNow() {
	f.failNowed = true
}

func (f *fakeTestingT) Fail() {
	f.failed = true
}

func (f *fakeTestingT) Log(args ...interface{}) {
	f.msgs = append(f.msgs, args[0].(string))
}

func (f *fakeTestingT) Helper() {}

func TestAssertWithBoolFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	Assert(fakeT, 1 == 6)
	expectFailNowed(t, fakeT, "assertion failed: 1 == 6 is false")
}

func TestAssertWithBoolFailureAndExtraMessage(t *testing.T) {
	fakeT := &fakeTestingT{}

	Assert(fakeT, 1 > 5, "sometimes things fail")
	expectFailNowed(t, fakeT, "assertion failed: 1 > 5 is false: sometimes things fail")
}

func TestAssertWithBoolSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}

	Assert(fakeT, 1 < 5)
	expectSuccess(t, fakeT)
}

func TestAssertWithBoolMultiLineFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	Assert(fakeT, func() bool {
		for range []int{1, 2, 3, 4} {
		}
		return false
	}())
	expectFailNowed(t, fakeT, `assertion failed: func() bool {
	for range []int{1, 2, 3, 4} {
	}
	return false
}() is false`)
}

type exampleComparison struct {
	success bool
	message string
}

func (c exampleComparison) Compare() (bool, string) {
	return c.success, c.message
}

func TestAssertWithComparisonSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}

	cmp := exampleComparison{success: true}
	Assert(fakeT, cmp.Compare)
	expectSuccess(t, fakeT)
}

func TestAssertWithComparisonFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	cmp := exampleComparison{message: "oops, not good"}
	Assert(fakeT, cmp.Compare)
	expectFailNowed(t, fakeT, "assertion failed: oops, not good")
}

func TestAssertWithComparisonAndExtraMessage(t *testing.T) {
	fakeT := &fakeTestingT{}

	cmp := exampleComparison{message: "oops, not good"}
	Assert(fakeT, cmp.Compare, "extra stuff %v", true)
	expectFailNowed(t, fakeT, "assertion failed: oops, not good: extra stuff true")
}

type customError struct{}

func (e *customError) Error() string {
	return "custom error"
}

func TestNilErrorSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}

	var err error
	NilError(fakeT, err)
	expectSuccess(t, fakeT)

	NilError(fakeT, nil)
	expectSuccess(t, fakeT)

	var customErr *customError
	NilError(fakeT, customErr)
}

func TestNilErrorFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	NilError(fakeT, fmt.Errorf("this is the error"))
	expectFailNowed(t, fakeT, "assertion failed: error is not nil: this is the error")
}

func TestCheckFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	if Check(fakeT, 1 == 2) {
		t.Error("expected check to return false on failure")
	}
	expectFailed(t, fakeT, "assertion failed: 1 == 2 is false")
}

func TestCheckSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}

	if !Check(fakeT, 1 == 1) {
		t.Error("expected check to return true on success")
	}
	expectSuccess(t, fakeT)
}

func TestEqualSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}

	Equal(fakeT, 1, 1)
	expectSuccess(t, fakeT)

	Equal(fakeT, "abcd", "abcd")
	expectSuccess(t, fakeT)
}

func TestEqualFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	Equal(fakeT, 1, 3)
	expectFailNowed(t, fakeT, "assertion failed: 1 (int) != 3 (int)")
}

func TestEqualFailureTypes(t *testing.T) {
	fakeT := &fakeTestingT{}

	Equal(fakeT, 3, uint(3))
	expectFailNowed(t, fakeT, `assertion failed: 3 (int) != 3 (uint)`)
}

type testingT interface {
	Errorf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})
}

func expectFailNowed(t testingT, fakeT *fakeTestingT, expected string) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if fakeT.failed {
		t.Errorf("should not have failed, got messages %s", fakeT.msgs)
	}
	if !fakeT.failNowed {
		t.Fatalf("should have failNowed with message %s", expected)
	}
	if fakeT.msgs[0] != expected {
		t.Fatalf("should have failure message %q, got %q", expected, fakeT.msgs[0])
	}
}

// nolint: unparam
func expectFailed(t testingT, fakeT *fakeTestingT, expected string) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if fakeT.failNowed {
		t.Errorf("should not have failNowed, got messages %s", fakeT.msgs)
	}
	if !fakeT.failed {
		t.Fatalf("should have failed with message %s", expected)
	}
	if fakeT.msgs[0] != expected {
		t.Fatalf("should have failure message %q, got %q", expected, fakeT.msgs[0])
	}
}

func expectSuccess(t testingT, fakeT *fakeTestingT) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if fakeT.failNowed {
		t.Errorf("should not have failNowed, got messages %s", fakeT.msgs)
	}
	if fakeT.failed {
		t.Errorf("should not have failed, got messages %s", fakeT.msgs)
	}
}