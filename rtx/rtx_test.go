package rtx

// The tests for this must be whitebox, because we have to override log.Fatal.

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
)

func success() error {
	return nil
}

func successValue(v int) (int, error) {
	return v, nil
}

func fail() error {
	return errors.New("A failure for testing")
}

func failValue(v int) (int, error) {
	return v, errors.New("A failure for testing")
}

func TestMustSuccess(t *testing.T) {
	Must(success(), "Works")
	i := ValueOrDie(successValue(5))
	if i != 5 {
		t.Errorf("ValueOrDie returned %d, not 5", i)
	}
}

type exiter struct {
	count int
}

func (e *exiter) dontExit(_ int) {
	e.count++
}

func TestMustFailure(t *testing.T) {
	// Inject our own output and failure routines.
	e := exiter{}
	osExit = e.dontExit
	defer func() { osExit = os.Exit }()

	defer log.SetOutput(os.Stdout)
	b := bytes.Buffer{}
	log.SetOutput(&b)

	// Call the function which causes the output
	Must(fail(), "Should fail")

	// Find out what got written.
	out := b.String()

	// Return logging back to normal state
	log.SetOutput(os.Stdout)

	if !strings.HasSuffix(string(out), "Should fail (error: A failure for testing)\n") {
		t.Errorf("%q does not end with \"Should fail (error: An error for testing)\"", out)
	}

	if e.count != 1 {
		t.Errorf("Should have exited once, not %d times", e.count)
	}
}

func TestValueOrDieFailure(t *testing.T) {
	// Inject our own output and failure routines.
	e := exiter{}
	osExit = e.dontExit
	defer func() { osExit = os.Exit }()

	defer log.SetOutput(os.Stdout)
	b := bytes.Buffer{}
	log.SetOutput(&b)

	_ = ValueOrDie(failValue(5))

	out := b.String()

	log.SetOutput(os.Stdout)

	if !strings.HasSuffix(out, "Fails (error: A failure for testing)\n") {
		t.Errorf("%q does not end with \"Fails (error: A failure for testing)\"", out)
	}

	if e.count != 1 {
		t.Errorf("Should have exited once, not %d times", e.count)
	}
}

func TestMustFailureWithFormatting(t *testing.T) {
	// Inject our own output and failure routines.
	e := exiter{}
	osExit = e.dontExit
	defer func() { osExit = os.Exit }()

	defer log.SetOutput(os.Stdout)
	b := bytes.Buffer{}
	log.SetOutput(&b)

	// Call the function which causes the output
	Must(fail(), "Should fail with arg %d", 5)

	// Find out what got read
	out := b.String()

	// Return logging back to normal state
	log.SetOutput(os.Stdout)

	if !strings.HasSuffix(out, "Should fail with arg 5 (error: A failure for testing)\n") {
		t.Errorf("%q does not end with \"Should fail with arg 5 (error: An error for testing)\"", out)
	}

	if e.count != 1 {
		t.Errorf("Should have exited once, not %d times", e.count)
	}
}

func TestPanicOnErrorWontPanicOnNil(t *testing.T) {
	PanicOnError(nil, "This should be fine")
	v := ValueOrPanic(successValue(8))
	if v != 8 {
		t.Errorf("ValueOrPanic returned %d, not 8", v)
	}
}

func TestPanicOnErrorPanicsOnError(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("We should have recovered from a panic")
		}
		if r != "Expect an error (error: Error for testing)" {
			t.Error(r, "was not the expected string")
		}
	}()
	PanicOnError(errors.New("Error for testing"), "Expect an error")
}

func TestValueOrPanicPanicsOnError(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("We should have recovered from a panic")
		}
		if r != "Expect an error (error: Error for testing)" {
			t.Error(r, "was not the expected string")
		}
	}()
	ValueOrPanic(failValue(8))
}

func TestPanicOnErrorPanicsOnErrorAndFormatsCorrectly(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("We should have recovered from a panic")
		}
		if r != "Expect an error and 1 should be one (error: Error for testing)" {
			t.Error(r, "was not the expected string")
		}
	}()
	PanicOnError(errors.New("Error for testing"), "Expect an error and %d should be one", 1)
}
