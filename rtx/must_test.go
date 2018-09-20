package rtx

// The tests for this must be whitebox, because we have to override log.Fatal.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func success() error {
	return nil
}

func fail() error {
	return errors.New("A failure for testing")
}

func TestMustSuccess(t *testing.T) {
	Must(success(), "Works")
}

func callPrintln(args ...interface{}) {
	fmt.Println(args...)
}

func TestMustFailure(t *testing.T) {
	logFatal = callPrintln
	defer func() { logFatal = log.Fatal }()

	// Technique from https://stackoverflow.com/questions/10473800/in-go-how-do-i-capture-stdout-of-a-function-into-a-string
	// Intercept stdout
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)

	// Copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// Call the function which causes the output
	Must(fail(), "Should fail")

	// Return back to normal state
	w.Close()
	os.Stdout = old

	// Find out what got read
	out := <-outC

	if string(out) != "Should fail (error: A failure for testing)\n" {
		t.Errorf("%q != \"Should fail (error: An error for testing)\"", out)
	}
}

func TestMustFailureWithFormatting(t *testing.T) {
	logFatal = callPrintln
	defer func() { logFatal = log.Fatal }()

	// Technique from https://stackoverflow.com/questions/10473800/in-go-how-do-i-capture-stdout-of-a-function-into-a-string
	// Intercept stdout
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)

	// Copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// Call the function which causes the output
	Must(fail(), "Should fail with arg %d", 5)

	// Return back to normal state
	w.Close()
	os.Stdout = old

	// Find out what got read
	out := <-outC

	if string(out) != "Should fail with arg 5 (error: A failure for testing)\n" {
		t.Errorf("%q != \"Should fail with arg 5 (error: An error for testing)\"", out)
	}
}
